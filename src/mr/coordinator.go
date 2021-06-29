package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"
import "sync"

//task's status----------------------------------
//UnAllocated represents unmapped
//Allocated represents it have been mapped to a worker
//Finished represents worker finish the map task
const (
	UnAllocated = iota
	Allocated
	Finished
)

//task chan
var maptasks chan string      // chan for map task
var reducetasks chan int      // chan for reduce task
//---------------------------------------------------

type Coordinator struct {
	// Your definitions here.
	AllFilesName     map[string]int     // splited files, map task's staus, where key is the file name,val is teh file's status
	MapTaskNumCount  int                // current map task number
	MapFinshed       bool               // finish the map task
	InterFilename    [][]string         // intermediate file

	NReduce          int                // n Reduce task
	ReduceTaskStatus map[int]int        // reduce task's status
	ReduceFinshed    bool               // finish the reduce task

	RWLock           *sync.RWMutex
	//--------------------------------------------
}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//generateTask:create tasks
func (c *Coordinator) generateTask(){
	for k,v :=range c.AllFilesName{
		if v==UnAllocated{
			maptasks <- k
		}
	}
	//check if all map tasks have finished
	ok :=false
	for !ok{
		ok = checkAllMapTask(c)
	}
	c.MapFinshed = true
	//check if all reduce tasks have finished
	for k,v := range c.ReduceTaskStatus{
		if v== UnAllocated{
			reducetasks <- k
		}
	}
    ok = false
    for !ok{
    	ok = checkAllReduceTask(c)
	}
	c.ReduceFinshed = true
}

//checkAllMapTask : check if all map tasks have finished
func checkAllMapTask(c *Coordinator) bool{
	c.RWLock.RLock()
	defer c.RWLock.RUnlock()
	for _,v := range c.AllFilesName{
		if v !=Finished{
			return false
		}
	}
	return true
}

//checkAllReduceTask: check if all reduce tasks have finished
func checkAllReduceTask(c *Coordinator) bool{
	c.RWLock.RLock()
	defer c.RWLock.RUnlock()
	for _,v := range c.ReduceTaskStatus{
		if v != Finished{
			return false
		}
	}
	return true
}
//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.
    ret = c.ReduceFinshed

	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.AllFilesName=make(map[string]int)
	c.MapTaskNumCount=0
	c.ReduceFinshed=false
    c.InterFilename=make([][]string,c.NReduce)

	c.NReduce=nReduce
	c.ReduceTaskStatus=make(map[int]int)
	c.ReduceFinshed=false
	c.RWLock=new(sync.RWMutex)

	for _,v :=range files{
		c.AllFilesName[v]=UnAllocated
	}

	for i:=0;i<nReduce;i++{
		c.ReduceTaskStatus[i]=UnAllocated
	}
	//----------------------------------------------------

	c.server()
	return &c
}
