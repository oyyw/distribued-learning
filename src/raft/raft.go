package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"math/rand"
	"sort"

	//"bytes"
	"sync"
	"sync/atomic"
	"time"
	//"awesomeProject/6.824/src/labgob"
	"awesomeProject/6.824/src/labrpc"
)


//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
//
type ApplyMsg struct {
	CommandValid bool
	//向application层提交日志
	Command      interface{}
	CommandIndex int
	CommandTerm  int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}
type LogEntry struct {
	Command interface{}
	Term    int
}

// 当前角色

const ROLE_LEADER = "Leader"
const ROLE_FOLLOWER = "Follower"
const ROLE_CANDIDATES = "Candidates"
//
// A Go object implementing a single Raft peer.
//

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm int         // 见过的最大任期
	votedFor    int         // 记录在currentTerm任期投票给谁了
	log         []LogEntry // 操作日志

	// 所有服务器，易失状态
	commitIndex int // 已知的最大已提交索引
	lastApplied int // 当前应用到状态机的索引

	// 仅Leader，易失状态（成为leader时重置）
	nextIndex  []int //	每个follower的log同步起点索引（初始为leader log的最后一项）
	matchIndex []int // 每个follower的log同步进度（初始为0），和nextIndex强关联

	// 所有服务器，选举相关状态
	role              string    // 身份
	leaderId          int       // leader的id
	lastActiveTime    time.Time // 上次活跃时间（刷新时机：收到leader心跳、给其他candidates投票、请求其他节点投票）
	lastBroadcastTime time.Time // 作为leader，上次的广播时间

	applyCh chan ApplyMsg
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
    rf.mu.Lock()
    defer rf.mu.Unlock()
	var term int
	var isleader bool
	// Your code here (2A).
	term = rf.currentTerm
	isleader = rf.role == ROLE_LEADER
	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//--
func (rf *Raft) persist() {
	// Your code here (2C).
	//
	// e := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}


//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}


//
// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
//
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}


//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}
type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term          int
	Success       bool
	ConflictTerm  int
	ConflictIndex int
}
//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term = rf.currentTerm
	reply.VoteGranted = false

	DPrintf("RaftNode[%d] Handle RequestVote, CandidatesId[%d] Term[%d] CurrentTerm[%d] LastLogIndex[%d] LastLogTerm[%d] votedFor[%d]",
		rf.me, args.CandidateId, args.Term, rf.currentTerm, args.LastLogIndex, args.LastLogTerm, rf.votedFor)
	defer func() {
		DPrintf("RaftNode[%d] Return RequestVote, CandidatesId[%d] VoteGranted[%v] ", rf.me, args.CandidateId, reply.VoteGranted)
	}()
	// 任期不如我大，拒绝投票
	if args.Term < rf.currentTerm {
		return
	}
	// 发现更大的任期，则转为该任期的follower
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.role = ROLE_FOLLOWER
		rf.votedFor = -1
		rf.leaderId = -1
		// 继续向下走，进行投票
	}

	// 每个任期，只能投票给1人
	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		// candidate的日志必须比我的新
		// 1, 最后一条log，任期大的更新
		// 2，任期相同, 更长的log则更新
		lastLogTerm := 0
		if len(rf.log) != 0 {
			lastLogTerm = rf.log[len(rf.log)-1].Term
		}
		// 这里坑了好久，一定要严格遵守论文的逻辑，另外log长度一样也是可以给对方投票的
		if args.LastLogTerm > lastLogTerm || (args.LastLogTerm == lastLogTerm && args.LastLogIndex >= len(rf.log)) {
			rf.votedFor = args.CandidateId
			reply.VoteGranted = true
			rf.lastActiveTime = time.Now() // 为其他人投票，那么重置自己的下次投票时间
		}
	}
	rf.persist()
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	DPrintf("RaftNode[%d] Handle AppendEntries, LeaderId[%d] Term[%d] CurrentTerm[%d] role=[%s]",
		rf.me, args.LeaderId, args.Term, rf.currentTerm, rf.role)
	defer func() {
		DPrintf("RaftNode[%d] Return AppendEntries, LeaderId[%d] Term[%d] CurrentTerm[%d] role=[%s]",
			rf.me, args.LeaderId, args.Term, rf.currentTerm, rf.role)
	}()

	reply.Term = rf.currentTerm
	reply.Success = false
	reply.ConflictIndex = -1
	reply.ConflictTerm = -1

	if args.Term < rf.currentTerm {
		return
	}

	// 发现更大的任期，则转为该任期的follower
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.role = ROLE_FOLLOWER
		rf.votedFor = -1
		rf.leaderId = -1
		rf.persist()
		// 继续向下走
	}

	// 认识新的leader
	rf.leaderId = args.LeaderId
	// 刷新活跃时间
	rf.lastActiveTime = time.Now()

	//------------------------------2B,日志--------------------------
	//如果本地没有前一个日志的话，那么false
	if len(rf.log) < args.PrevLogIndex{
		//*******6.824_2B上更改
		reply.ConflictIndex = len(rf.log) + 1
		//***********************
		return
	}
	//如果本地有前一个日志的话，那么term必须相同，否则false、
	if args.PrevLogIndex > 0 && rf.log[args.PrevLogIndex - 1].Term != args.PrevLogTerm{
		//---------------------3---------------------------
		reply.ConflictTerm = rf.log[args.PrevLogIndex - 1].Term
		for index := 2; index <= args.PrevLogIndex;index ++{ //找到冲突term的首次出现位置，最差是PrevLogIndex
			if rf.log[index-1].Term == reply.ConflictTerm{
				reply.ConflictIndex = index
				break
			}
		}
		//---------------------------------------------------
		return
	}
	//保持日志
	for i, logEntry :=range args.Entries{
		index :=args.PrevLogIndex +i+1
		//-------4---------
		logPos := index -1
		if index > len(rf.log){
			rf.log = append(rf.log, logEntry)
		}else{ //重叠部分
			if rf.log[logPos].Term !=logEntry.Term{
				rf.log = rf.log[:logPos]         //删除当前以及后续所有log
				rf.log = append(rf.log, logEntry)  //把新的log加入进来
			}  //term 一样啥也不用做，继续向后比对log
		}
	}

	// 日志操作lab-2A不实现
	rf.persist()


	//更新提交下标
	if args.LeaderCommit > rf.commitIndex{
		rf.commitIndex = args.LeaderCommit
		if len(rf.log) < rf.commitIndex{
			rf.commitIndex = len(rf.log)
		}
	}
	reply.Success = true
	//----------------------------------2B--------------------------------

}
//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) electionLoop() {
	for !rf.killed() {
		time.Sleep(10*time.Millisecond)

		func() {
			rf.mu.Lock()
			defer rf.mu.Unlock()
			now := time.Now()
			timeout := time.Duration(300+rand.Int31n(150)) * time.Millisecond // 超时随机化
			elapses := now.Sub(rf.lastActiveTime)
			// follower -> candidates
			if rf.role == ROLE_FOLLOWER {
				if elapses >= timeout {
					rf.role = ROLE_CANDIDATES
					DPrintf("RaftNode[%d] Follower -> Candidate", rf.me)
				}
			}
			// 请求vote
			if rf.role == ROLE_CANDIDATES && elapses >= timeout {
				rf.lastActiveTime = now // 重置下次选举时间

				rf.currentTerm += 1 // 发起新任期
				rf.votedFor = rf.me // 该任期投了自己
				rf.persist()

				// 请求投票req
				args := RequestVoteArgs{
					Term:         rf.currentTerm,
					CandidateId:  rf.me,
					LastLogIndex: len(rf.log),
				}
				if len(rf.log) != 0 {
					args.LastLogTerm = rf.log[len(rf.log)-1].Term
				}

				rf.mu.Unlock()

//				fmt.Printf("RaftNode[%d] RequestVote starts, Term[%d] LastLogIndex[%d] LastLogTerm[%d]\n", rf.me, args.Term,
//					args.LastLogIndex, args.LastLogTerm)

				// 并发RPC请求vote
				type VoteResult struct {
					peerId int
					resp   *RequestVoteReply
				}
				voteCount := 1   // 收到投票个数（先给自己投1票）
				finishCount := 1 // 收到应答个数
				voteResultChan := make(chan *VoteResult, len(rf.peers))
				for peerId := 0; peerId < len(rf.peers); peerId++ {
//					fmt.Printf("向节点[%d]请求投票",peerId)
					go func(id int) {
						if id == rf.me {
							return
						}
						resp := RequestVoteReply{}
						if ok := rf.sendRequestVote(id, &args, &resp); ok {
							voteResultChan <- &VoteResult{peerId: id, resp: &resp}
						} else {
							voteResultChan <- &VoteResult{peerId: id, resp: nil}
						}
					}(peerId)
				}
				maxTerm := 0
				for {
					select {
					case voteResult := <-voteResultChan:
						finishCount += 1
//						fmt.Println("收到投票结果")
						if voteResult.resp != nil {
							if voteResult.resp.VoteGranted {
								voteCount += 1
							}
							if voteResult.resp.Term > maxTerm {
								maxTerm = voteResult.resp.Term
							}
						}
						// 得到大多数vote后，立即离开
						if finishCount == len(rf.peers) || voteCount > len(rf.peers)/2 {
							goto VOTE_END
						}
					}
				}
			VOTE_END:
				rf.mu.Lock()
				defer func() {
//					fmt.Printf("RaftNode[%d] RequestVote ends, finishCount[%d] voteCount[%d] Role[%s] maxTerm[%d] currentTerm[%d]\n", rf.me, finishCount, voteCount,
//						rf.role, maxTerm, rf.currentTerm)
				}()
				// 如果角色改变了，则忽略本轮投票结果
				if rf.role != ROLE_CANDIDATES {
					return
				}
				// 发现了更高的任期，切回follower
				if maxTerm > rf.currentTerm {
					rf.role = ROLE_FOLLOWER
					rf.leaderId = -1
					rf.currentTerm = maxTerm
					rf.votedFor = -1
					rf.persist()
					return
				}
				// 赢得大多数选票，则成为leader
				if voteCount > len(rf.peers)/2 {
					rf.role = ROLE_LEADER
					rf.leaderId = rf.me

					//---------------2B------------------
					rf.nextIndex = make([]int,len(rf.peers))
					for i:=0;i<len(rf.peers);i++{
						rf.nextIndex[i] = len(rf.log)+1
					}
					rf.matchIndex = make([]int,len(rf.peers))
					for i:=0; i<len(rf.peers);i++{
					    rf.matchIndex[i] = 0
					}
                    //----------------2B---------------------

					rf.lastBroadcastTime = time.Unix(0, 0) // 令appendEntries广播立即执行
					return
				}
			}
		}()
	}
}

// lab-2A只做心跳，不考虑log同步
func (rf *Raft) appendEntriesLoop() {
	for !rf.killed() {
		time.Sleep(10 * time.Millisecond)
		func() {
			rf.mu.Lock()
			defer rf.mu.Unlock()
			//------------------------------2A--------------------------
			// 只有leader才向外广播心跳
			if rf.role != ROLE_LEADER {
				return
			}

			// 100ms广播1次
			now := time.Now()
			if now.Sub(rf.lastBroadcastTime) < 100*time.Millisecond {
				return
			}
			rf.lastBroadcastTime = time.Now()

			// 并发RPC心跳
			type AppendResult struct {
				peerId int
				resp   *AppendEntriesReply
			}

			for peerId := 0; peerId < len(rf.peers); peerId++ {
				if peerId == rf.me {
					continue
				}
				args := AppendEntriesArgs{}
				args.Term = rf.currentTerm
				args.LeaderId = rf.me
				// log相关字段在lab-2A不处理
				//-----------------------------2B-------------------------------------

				args.LeaderCommit = rf.commitIndex
				args.Entries = make([]LogEntry, 0)
				args.PrevLogIndex = rf.nextIndex[peerId] - 1
				if args.PrevLogIndex > 0 {
					args.PrevLogTerm = rf.log[args.PrevLogIndex - 1].Term
				}
			//	args.Entries = append(args.Entries, rf.log[rf.nextIndex[peerId]-1:]...)
			    args.Entries = append(args.Entries,rf.log[args.PrevLogIndex:]...)
				DPrintf("RaftNode[%d] appendEntries starts, currentTerm[%d] peer[%d] logIndex=[%d] nextIndex[%d] matchIndex[%d] args.Entries[%d] commitIndex[%d]",
					rf.me, rf.currentTerm, peerId, len(rf.log), rf.nextIndex[peerId], rf.matchIndex[peerId],len(args.Entries),rf.commitIndex)

				//-----------------------------2B-------------------------------------

				go func(id int, args1 *AppendEntriesArgs) {
					//RPC期间无锁
					reply := AppendEntriesReply{}
					if ok := rf.sendAppendEntries(id, args1, &reply); ok {
						rf.mu.Lock()
						defer rf.mu.Unlock()

						//如果不是RPC之前的leader状态了，就什么也不做
						if rf.currentTerm != args1.Term{
							return
						}
						// 发现更大的任期，变成follower
						if reply.Term > rf.currentTerm {
							rf.role = ROLE_FOLLOWER
							rf.leaderId = -1
							rf.currentTerm = reply.Term
							rf.votedFor = -1
							rf.persist()
							return
						}

						//------------------2B-----------------------------

						if reply.Success{
							// 因为RPC期间无锁, 可能相关状态被其他RPC修改了
							// 因此这里得根据发出RPC请求时的状态做更新，而不要直接对nextIndex和matchIndex做相对加减
							// 因为success表示已成功进行log replication,这时应该在原index(args1.preLogIndex + len(args1.Entries))
							// 的基础上再加1
							rf.nextIndex[id] = args1.PrevLogIndex + len(args1.Entries) + 1
					       //---------1.----------  rf.nextIndex[id] += len(args1.Entries)
							rf.matchIndex[id] = rf.nextIndex[id] -1

							// 数字N, 让nextIndex[i]的大多数>=N
							// peer[0]' index=2
							// peer[1]' index=2
							// peer[2]' index=1
							// 1,2,2
							// 更新commitIndex, 就是找中位数
							sortedMatchIndex := make([]int,0)
							sortedMatchIndex = append(sortedMatchIndex, len(rf.log))
							for i:=0; i<len(rf.peers); i++{
								if i == rf.me{
									continue
								}
								sortedMatchIndex = append(sortedMatchIndex, rf.matchIndex[i])
							}
							sort.Ints(sortedMatchIndex)
					//		fmt.Println(sortedMatchIndex)
							newCommitIndex := sortedMatchIndex[len(rf.peers)/2]
						//	fmt.Println(rf.log)
						//	fmt.Println(newCommitIndex)
							//2----------  if newCommitIndex > rf.commitIndex && rf.log[newCommitIndex -1].Term == rf.currentTerm{
							if newCommitIndex>rf.commitIndex && rf.log[newCommitIndex -1].Term == rf.currentTerm{
								rf.commitIndex = newCommitIndex
							}
						}else{
							//回退优化 参考：https://thesquareplanet.com/blog/students-guide-to-raft/#an-aside-on-optimizations
						/*	rf.nextIndex[id] -=1
							if rf.nextIndex[id] < 1{
								rf.nextIndex[id] = 1
							}*/
							// -----5---------------------
							if reply.ConflictTerm != -1 { // follower的prevLogIndex位置term冲突了
								/*
								* 这时有两种情况: 1. leader中无这个冲突term，此情形，follower的整个conflictTerm均得删除
								*                  所以，我们可以设置conflictIndex为此follower的 conflictTerm的第一个index
								*               2. leader中有这个冲突Term，此情形，follower的这个conflictTerm中不一定所有的index都冲突
								*                  所以, 我们可以设置conflictTerm为 leader log中confictTerm最后出现的位置
								*/
								conflictTermIndex := -1
								for index := args.PrevLogIndex; index > 1; index-- {
									if rf.log[index -1 ].Term == reply.ConflictTerm {
										conflictTermIndex = index
										break
									}
								}
								if conflictTermIndex != -1 { // leader log出现了这个term，那么从这里prevLogIndex之前的最晚出现位置尝试同步
									rf.nextIndex[id] = conflictTermIndex
								//	rf.nextIndex[id] = 1
								} else {
									rf.nextIndex[id] = reply.ConflictIndex // 用follower首次出现term的index作为同步开始
								}
							} else {
								// follower没有发现prevLogIndex term冲突, 可能是被snapshot了或者日志长度不够
								// 这时候我们将返回的conflictIndex设置为nextIndex即可
								//********6.824_2B上更改**************
								rf.nextIndex[id] = reply.ConflictIndex
								//*********************
							//*6.824_2B*****	rf.nextIndex[id] = 1
							}
						}
						//------------------------2B-------------------------
						DPrintf("RaftNode[%d] appendEntries ends,  currentTerm[%d]  peer[%d] logIndex=[%d] nextIndex[%d] matchIndex[%d] commitIndex[%d]",
							rf.me, rf.currentTerm, id, len(rf.log), rf.nextIndex[id], rf.matchIndex[id], rf.commitIndex)
					}
				}(peerId, &args)
			}
		}()
	}
}
//心跳


//日志提交--------------------2B------------------------
func (rf *Raft) applyLogLoop(applyCh chan ApplyMsg){
	noMore := false
	for !rf.killed(){
		if noMore{
			time.Sleep(10 * time.Millisecond)
		}
		var appliedMsgs = make([]ApplyMsg, 0)

		func(){
			rf.mu.Lock()
			defer rf.mu.Unlock()
            noMore = true

			for rf.commitIndex >rf.lastApplied{
				rf.lastApplied +=1
				appliedMsgs = append(appliedMsgs,ApplyMsg{
					CommandValid: true,
					Command:      rf.log[rf.lastApplied - 1].Command,
					CommandIndex: rf.lastApplied,
					CommandTerm:  rf.log[rf.lastApplied - 1].Term,
				})
				DPrintf("RaftNode[%d] applylog, currentTerm[%d] lastApplide[%d] commitIndex[%d]",
					rf.me, rf.currentTerm,rf.lastApplied,rf.commitIndex)
			}
		}()
		for _, msg :=range appliedMsgs{
			applyCh <- msg
		}
	}
}
//-----------------------------2B-----------------------------


//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//


//---------------------2b--------------------------
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).
    rf.mu.Lock()
	defer rf.mu.Unlock()
	//只有leader才能写入
	if rf.role !=ROLE_LEADER{
		return -1,-1,false
	}
	logEntry :=LogEntry{
			Command: command,
			Term: rf.currentTerm,
	}
	rf.log = append(rf.log, logEntry)
	index = len(rf.log)
	term = rf.currentTerm
	rf.persist()

	DPrintf("RaftNode[%d] Add Command, logIndex[%d] currentTerm[%d]",rf.me,index,term)
	return index, term, isLeader
}
//------------------------------2B-----------------------------


//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}
/*
// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for rf.killed() == false {

		// Your code here to check if a leader election should
		// be started and to randomize sleeping time using
		// time.Sleep().

	}
}*/

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.role = ROLE_FOLLOWER
	rf.leaderId = -1
	rf.votedFor = -1
	rf.lastActiveTime = time.Now()
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
//	go rf.ticker()
	// election逻辑
	go rf.electionLoop()
	// leader逻辑
	go rf.appendEntriesLoop()
    //apply 逻辑
	go rf.applyLogLoop(applyCh)
//	fmt.Printf("Raftnode[%d]启动", me)

	return rf
}
