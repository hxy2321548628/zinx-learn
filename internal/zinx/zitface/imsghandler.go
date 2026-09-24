package zitface

/*
消息管理抽象层
*/
type IMsgHandle interface {
	DoMsgHandler(request IRequest)          // 根据消息 ID 调用对应的路由。
	AddRouter(msgId uint32, router IRouter) //为消息添加具体的处理逻辑
}
