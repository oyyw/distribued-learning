# distribued-learning，此仓库用于学习、管理分布式

## 一. Raft协议

* 此协议代码在 src/Raft/raft.go

#### 一. 对于Raft的一些思考

* Raft协议如何保证不删除已提交日志和避免提交前任未提交日志
<https://blog.csdn.net/ouyangyiwen/article/details/119922657>

* 为什么持久化存储的数据为log、currentTerm和votefor
<https://blog.csdn.net/ouyangyiwen/article/details/119902201>

#### 二. 而关于代码实现问题

* Raft协议问题总结(一): Leader 选举的异步选举与提前决策
<https://blog.csdn.net/ouyangyiwen/article/details/119902185>

* Raft协议问题总结(二): 冲突Term和Index的回退优化
<https://blog.csdn.net/ouyangyiwen/article/details/119902194>

* Log peisist 实现
<https://blog.csdn.net/ouyangyiwen/article/details/119922672>
  
#### Raft处理重复请求
  - 每个要做proposal的client需要一个唯一的identifier,它的每个不同的proposal需要有一个顺序递增的序列 
    号，cliend id 和这个序列号由此可以唯一确定一个不同的proposal，从而使得各个raft节点可以记录保存各proposal应用以后的结果。
  - 当一个proposal超时，client不提高proposal的序列号，使用原proposal序列号重试。 
  - 当一个proposal被成功提交并应用且被成功回复给client以后，client顺序提高proposal的序列号，并记录下收到的成功回复的proposal的序列号。
    raft节点收到一个proposal请求以后，得到请求中夹带的这个最大成功回复的proposal的序列号，它和它之前所有的应用结果都可以删去。
    proposal序列号和client id可用于判断这个proposal是否应用过，如果已经应用过，则不再再次应用，直接返回已保存的结果。
    等于是每个不同的proposal可以被commit多次，在log中出现多次，但永远只会被apply一次。
  - 系统维护一定数量允许的client数量，比如可以用LRU策略淘汰。请求过来了，而client已经被LRU淘汰掉了，则让client直接fail掉。
  - 这些已经注册的client信息，包括和这些client配套的上述proposal结果、各序列号等等，需要在raft组内一致的维护。
    也就是说，上述各raft端数据结构和它们的操作实际是state machine的一部分。在做snapshotting的时候，它们同样需要被保存与恢复。
