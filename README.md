# distribued-learning，此仓库用于学习、管理分布式

## Raft协议

* 此协议代码在 src/Raft/raft.go

### 一. 对于Raft的一些思考

#### 1. Raft协议如何保证不删除已提交日志和避免提交前任未提交日志
<https://blog.csdn.net/ouyangyiwen/article/details/119922657>

#### 2. 为什么持久化存储的数据为log、currentTerm和votefor
<https://blog.csdn.net/ouyangyiwen/article/details/119902201>

### 二. 而关于代码实现问题

#### Raft协议问题总结(一): Leader 选举的异步选举与提前决策
<https://blog.csdn.net/ouyangyiwen/article/details/119902185>

#### Raft协议问题总结(二): 冲突Term和Index的回退优化
<https://blog.csdn.net/ouyangyiwen/article/details/119902194>

#### Log peisist 实现
<https://blog.csdn.net/ouyangyiwen/article/details/119922672>