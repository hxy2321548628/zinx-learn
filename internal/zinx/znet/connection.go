package znet

import (
	"fmt"
	"net"
	"zinx-learn/internal/zinx/zitface"
)

// Connection 封装一个客户端 TCP 连接及其关联的业务路由。
type Connection struct {
	Conn         *net.TCPConn    // Conn 是与客户端建立的底层 TCP 套接字。
	ConnID       string          // ConnID 是连接的全局唯一标识，也可视为会话 ID。
	isClosed     bool            // isClosed 标记连接是否已经关闭。
	ExitBuffChan chan bool       // ExitBuffChan 用于通知 Start 结束阻塞并退出连接。
	Router       zitface.IRouter // Router 处理从当前连接读取到的请求。
}

// GetTCPConnection 返回底层 TCP 连接。
func (this *Connection) GetTCPConnection() *net.TCPConn {
	return this.Conn
}

// GetConnID 返回连接的全局唯一标识。
func (this *Connection) GetConnID() string {
	return this.ConnID
}

// RemoteAddr 返回客户端的网络地址。
func (this *Connection) RemoteAddr() net.Addr {
	return this.Conn.RemoteAddr()
}

// Stop 关闭底层套接字，并通知 Start 结束等待。
func (this *Connection) Stop() {
	// 二次校验, 防止并发冲突
	//1. 如果当前链接已经关闭
	if this.isClosed == true {
		return
	}
	this.isClosed = true

	//TODO Connection Stop() 如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用

	// 关闭socket链接
	this.Conn.Close()

	//通知从缓冲队列读数据的业务，该链接已经关闭
	this.ExitBuffChan <- true

	//关闭该链接全部管道
	close(this.ExitBuffChan)
}

// StartReader 持续读取客户端数据，将每次读取封装成 Request 后交给路由处理。
func (this *Connection) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(this.RemoteAddr().String(), " conn reader exit!")
	defer this.Stop()

	for {
		// 当前版本尚未实现消息封包，先使用固定缓冲区演示一次读取对应一次请求。
		buf := make([]byte, 512)
		n, err := this.Conn.Read(buf)
		if err != nil {
			fmt.Println("recv buf err ", err)
			return
		}

		// 只传递本次实际读取的 n 个字节，避免把缓冲区尾部的零值交给业务层。
		req := Request{
			conn: this,
			data: buf[:n],
		}

		// 每个请求独立执行路由流程，三个处理阶段在同一 goroutine 内保持先后顺序。
		go func(request zitface.IRequest) {
			this.Router.PreHandle(request)
			this.Router.Handle(request)
			this.Router.PostHandle(request)
		}(&req)
	}
}

// Start 启动读协程，并阻塞到连接收到退出通知。
func (this *Connection) Start() {

	//开启处理该链接读取到客户端数据之后的请求业务
	go this.StartReader()

	for {
		select {
		case <-this.ExitBuffChan:
			//得到退出消息，不再阻塞
			return
		}
	}
}

// NewConntion 创建连接对象，并将服务器注册的路由绑定到该连接。
func NewConntion(
	conn *net.TCPConn,
	connID string,
	router zitface.IRouter) zitface.IConnection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		Router:       router,
		ExitBuffChan: make(chan bool, 1),
	}

	return c
}
