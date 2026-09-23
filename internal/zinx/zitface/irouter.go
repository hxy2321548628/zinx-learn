package zitface

// IRouter 定义一次请求依次经过的三个处理阶段。
// IRequest 同时携带连接与请求数据，框架使用同一个请求对象按顺序调用三个钩子。
type IRouter interface {
	PreHandle(IRequest)  // PreHandle 在核心业务执行前调用，可用于预处理。
	Handle(IRequest)     // Handle 执行请求的核心业务逻辑。
	PostHandle(IRequest) // PostHandle 在核心业务执行后调用，可用于收尾处理。
}
