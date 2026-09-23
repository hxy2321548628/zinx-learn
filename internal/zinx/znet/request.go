package znet

import "zinx-learn/internal/zinx/zitface"

// Request 是一次客户端请求的默认实现。
// conn 标识请求来源，data 只保存本次 Read 实际读取到的有效字节。
type Request struct {
	conn zitface.IConnection
	data []byte
}

// GetConnection 返回产生当前请求的客户端连接。
func (r *Request) GetConnection() zitface.IConnection {
	return r.conn
}

// GetData 返回本次请求携带的有效数据。
func (r *Request) GetData() []byte {
	return r.data
}
