package zitface

/*
消息管理抽象层
*/
type IMsgHandle interface {
	DoMsgHandler(request IRequest)          // 根据消息 ID 调用对应的路由。
	AddRouter(msgId uint32, router IRouter) //为消息添加具体的处理逻辑
	StartWorkerPool()                       //启动worker工作池
	SendMsgToTaskQueue(request IRequest)    //将消息交给TaskQueue,由worker进行处理
	GetWorkerPoolSize() uint32
	GetMaxWorkerTaskLen() uint32
}
