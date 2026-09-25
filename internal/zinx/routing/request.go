package routing

// Request 将已解码的数据、连接标识与回复能力关联起来。
// 它只暴露路由层实际需要的信息，避免处理器依赖底层 net.Conn。
type Request struct {
	responder    Responder
	connectionID uint32
	messageID    uint32
	data         []byte
}

// NewRequest 创建一个已完整解码的请求。
// 请求的生命周期由传给 Handler.Handle 的 context 控制，因此 Request 本身不保存 context。
func NewRequest(responder Responder, connectionID, messageID uint32, data []byte) *Request {
	return &Request{
		responder:    responder,
		connectionID: connectionID,
		messageID:    messageID,
		data:         data,
	}
}

// Responder 返回当前请求的回复器。
func (r *Request) Responder() Responder {
	return r.responder
}

// ConnectionID 返回产生当前请求的连接 ID。
func (r *Request) ConnectionID() uint32 {
	return r.connectionID
}

// Data 返回请求数据。
func (r *Request) Data() []byte {
	return r.data
}

// MessageID 返回请求的消息 ID。
func (r *Request) MessageID() uint32 {
	return r.messageID
}
