package znet

import (
	"fmt"
	"strconv"
	"zinx-learn/internal/zinx/zitface"
)

type MsgHandler struct {
	Apis             map[uint32]zitface.IRouter //存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize   uint32                     //业务工作Worker池的数量
	MaxWorkerTaskLen uint32                     // 每个队列存放最大的消息数量
	TaskQueue        []chan zitface.IRequest    //Worker负责取任务的消息队列
}

func NewMsgHandler(workerPoolSize uint32, maxWorkerTaskLen uint32) zitface.IMsgHandle {
	return &MsgHandler{
		Apis:             make(map[uint32]zitface.IRouter),
		WorkerPoolSize:   workerPoolSize,
		MaxWorkerTaskLen: maxWorkerTaskLen,
		TaskQueue:        make([]chan zitface.IRequest, workerPoolSize),
	}
}

func (mh *MsgHandler) GetWorkerPoolSize() uint32 {
	return mh.WorkerPoolSize
}
func (mh *MsgHandler) GetMaxWorkerTaskLen() uint32 {
	return mh.MaxWorkerTaskLen
}

// 启动一个Worker工作流程
func (mh *MsgHandler) StartOneWorker(workerID int, taskQueue chan zitface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started.")
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// 启动worker工作池
func (mh *MsgHandler) StartWorkerPool() {
	//遍历需要启动worker的数量，依此启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan zitface.IRequest, mh.MaxWorkerTaskLen)
		//启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 将消息交给TaskQueue,由worker进行处理
func (mh *MsgHandler) SendMsgToTaskQueue(request zitface.IRequest) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法则

	//得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnID=", request.GetConnection().GetConnID(), " request msgID=", request.GetMsgID(), "to workerID=", workerID)
	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}

func (mh *MsgHandler) DoMsgHandler(request zitface.IRequest) {
	router, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), " is not FOUND!")
		return
	}

	router.PreHandle(request)
	router.Handle(request)
	router.PostHandle(request)

}
func (mh *MsgHandler) AddRouter(msgId uint32, router zitface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	//2 添加msg与api的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}
