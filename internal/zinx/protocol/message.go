package protocol

// Message 是一条已完整解码的消息。
// 数据长度始终由 data 计算，避免长度字段与实际数据不一致。
type Message struct {
	id   uint32
	data []byte
}

// NewMessage 创建一条消息。
// data 不会被复制；调用方在消息处理完成前不应修改其底层字节切片。
func NewMessage(id uint32, data []byte) *Message {
	return &Message{id: id, data: data}
}

// ID 返回消息 ID。
func (m *Message) ID() uint32 {
	return m.id
}

// Data 返回消息体。
func (m *Message) Data() []byte {
	return m.data
}
