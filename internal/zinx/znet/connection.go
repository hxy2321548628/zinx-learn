package znet

import (
	"errors"
	"fmt"
	"net"
	"zinx-learn/internal/zinx/zitface"
)

type Connection struct {
	Conn         *net.TCPConn     //当前连接的socket TCP套接字
	ConnID       string           //当前连接的ID 也可以称作为SessionID，ID全局唯一
	isClosed     bool             //当前连接的关闭状态
	ExitBuffChan chan bool        //告知该链接已经退出/停止的channel
	handleAPI    zitface.HandFunc //该连接的处理方法api
}

// 从当前连接获取原始的socket TCPConn
func (this *Connection) GetTCPConnection() *net.TCPConn {
	return this.Conn
}

// 获取当前连接ID
func (this *Connection) GetConnID() string {
	return this.ConnID
}

// 获取远程客户端地址信息
func (this *Connection) RemoteAddr() net.Addr {
	return this.Conn.RemoteAddr()
}

// 停止连接，结束当前连接状态M
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

/* 处理conn读数据的Goroutine */
func (this *Connection) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(this.RemoteAddr().String(), " conn reader exit!")
	defer this.Stop()

	for {
		//读取我们最大的数据到buf中
		buf := make([]byte, 512)
		cnt, err := this.Conn.Read(buf)
		if err != nil {
			fmt.Println("recv buf err ", err)
			this.ExitBuffChan <- true
			continue
		}
		//调用当前链接业务(这里执行的是当前conn的绑定的handle方法)
		if err := this.handleAPI(this.Conn, buf, cnt); err != nil {
			fmt.Println("connID ", this.ConnID, " handle is error")
			this.ExitBuffChan <- true
			return
		}
	}
}

// 启动连接，让当前连接开始工作
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

// 创建连接的方法
func NewConntion(
	conn *net.TCPConn,
	connID string,
	callback_api zitface.HandFunc) zitface.IConnection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		handleAPI:    callback_api,
		ExitBuffChan: make(chan bool, 1),
	}

	return c
}

// 回显业务
// ============== 定义当前客户端链接的handle api ===========
func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
	//回显业务
	fmt.Println("[Conn Handle] CallBackToClient ... ")
	if _, err := conn.Write(data[:cnt]); err != nil {
		fmt.Println("write back buf err ", err)
		return errors.New("CallBackToClient error")
	}
	return nil
}
