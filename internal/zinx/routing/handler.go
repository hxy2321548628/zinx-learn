package routing

import "context"

// Responder 是处理器回复客户端所需的最小能力。
// 接口定义在使用方 routing 包中，使处理器不必依赖具体的 TCP 连接实现。
type Responder interface {
	SendMessage(ctx context.Context, messageID uint32, data []byte) error
}

// Handler 处理一条已完整解码的请求。
//
// ctx 在服务器关闭或连接结束时会被取消，耗时处理应定期检查它并尽快返回。
// 不同 worker 可能并发调用同一个 Handler；实现若持有可变状态，需要自行同步访问。
type Handler interface {
	Handle(context.Context, *Request) error
}

// HandlerFunc 允许普通函数直接作为 Handler 注册。
type HandlerFunc func(context.Context, *Request) error

// Handle 调用底层处理函数。
func (f HandlerFunc) Handle(ctx context.Context, request *Request) error {
	return f(ctx, request)
}
