package zitface

import "net"

type IConnection interface {
	//启动连接，让当前连接开始⼯工作
	Start()
	//停⽌止连接，结束当前连接状态M
	Stop()
	//从当前连接获取原始的socket TCPConn GetTCPConnection() *net.TCPConn
	// 获取当前连接ID
	GetConnID() string //获取远程客户端地址信息 RemoteAddr() net.Addr
}

// 定义一个统⼀处理理链接业务的接⼝口
// 第一参数是socket原⽣生链接
// 第⼆个参数是客户端请求的数据，
// 第三个参数是客户端请求的数据⻓长度。
type HandFunc func(*net.TCPConn, []byte, int) error
