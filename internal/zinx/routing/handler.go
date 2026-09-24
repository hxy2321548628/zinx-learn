package routing

// Responder 是处理器回复客户端所需的最小能力。
type Responder interface {
	SendMessage(messageID uint32, data []byte) error
}

// Handler 处理一条已完整解码的请求。
type Handler interface {
	Handle(*Request)
}

// HandlerFunc 允许普通函数直接作为 Handler 注册。
type HandlerFunc func(*Request)

// Handle 调用底层处理函数。
func (f HandlerFunc) Handle(request *Request) {
	f(request)
}
