package zitface

import "net"

// IConnection 定义框架对单个客户端连接提供的最小操作集合。
// 上层路由通过该接口访问连接，避免依赖 znet.Connection 的具体实现。
type IConnection interface {
	Start()                         // Start 启动连接的读协程并等待连接退出。
	Stop()                          // Stop 关闭套接字并结束当前连接。
	GetConnID() string              // GetConnID 返回当前连接的全局唯一标识。
	GetTCPConnection() *net.TCPConn // GetTCPConnection 返回底层 TCP 连接，供当前阶段的路由直接收发数据。
}
