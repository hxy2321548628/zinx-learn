package zitface

// IServer 定义 Zinx 服务器对外暴露的生命周期与路由注册能力。
type IServer interface {
	Start()                   // Start 异步启动网络监听。
	Stop()                    // Stop 停止服务器并释放相关资源。
	Serve()                   // Serve 启动服务器并阻塞调用方，使服务持续运行。
	AddRouter(router IRouter) // AddRouter 注册处理客户端请求的路由。
}
