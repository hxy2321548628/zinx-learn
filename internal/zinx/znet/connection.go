package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"zinx-learn/internal/zinx/zitface"
)

// Connection 封装一个客户端 TCP 连接及其关联的业务路由。
type Connection struct {
	Conn          *net.TCPConn       // Conn 是与客户端建立的底层 TCP 套接字。
	ConnID        string             // ConnID 是连接的全局唯一标识，也可视为会话 ID。
	isClosed      bool               // isClosed 标记连接是否已经关闭。
	ExitBuffChan  chan bool          // ExitBuffChan 用于通知 Start 结束阻塞并退出连接。
	MaxPacketSize uint32             // MaxPacketSize 限制单个数据包的最大字节数。
	MsgHandler    zitface.IMsgHandle //当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
}

// GetTCPConnection 返回底层 TCP 连接。
func (cn *Connection) GetTCPConnection() *net.TCPConn {
	return cn.Conn
}

// GetConnID 返回连接的全局唯一标识。
func (cn *Connection) GetConnID() string {
	return cn.ConnID
}

// RemoteAddr 返回客户端的网络地址。
func (cn *Connection) RemoteAddr() net.Addr {
	return cn.Conn.RemoteAddr()
}

// Stop 关闭底层套接字，并通知 Start 结束等待。
func (cn *Connection) Stop() {
	// 二次校验, 防止并发冲突
	//1. 如果当前链接已经关闭
	if cn.isClosed == true {
		return
	}
	cn.isClosed = true

	//TODO Connection Stop() 如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用

	// 关闭socket链接
	cn.Conn.Close()

	//通知从缓冲队列读数据的业务，该链接已经关闭
	cn.ExitBuffChan <- true

	//关闭该链接全部管道
	close(cn.ExitBuffChan)
}

// StartReader 持续读取客户端数据，将每次读取封装成 Request 后交给路由处理。
func (cn *Connection) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(cn.RemoteAddr().String(), " conn reader exit!")
	defer cn.Stop()

	for {
		// 创建拆包解包的对象
		dp := NewDataPack(cn.MaxPacketSize)

		//读取客户端的Msg head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(cn.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			cn.ExitBuffChan <- true
			continue
		}

		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			cn.ExitBuffChan <- true
			continue
		}

		//根据 dataLen 读取 data，放在 msg.Data 中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(cn.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				cn.ExitBuffChan <- true
				continue
			}
		}
		msg.SetData(data)

		// 只传递本次实际读取的 n 个字节，避免把缓冲区尾部的零值交给业务层。
		req := Request{
			conn: cn,
			msg:  msg,
		}

		// 每个请求独立执行路由流程，三个处理阶段在同一 goroutine 内保持先后顺序。
		go cn.MsgHandler.DoMsgHandler(&req)
	}
}

// Start 启动读协程，并阻塞到连接收到退出通知。
func (cn *Connection) Start() {

	//开启处理该链接读取到客户端数据之后的请求业务
	go cn.StartReader()

	for {
		select {
		case <-cn.ExitBuffChan:
			//得到退出消息，不再阻塞
			return
		}
	}
}

// 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	//将data封包，并且发送
	dp := NewDataPack(c.MaxPacketSize)
	msg, err := dp.Pack(NewMessage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	if _, err := c.Conn.Write(msg); err != nil {
		fmt.Println("Write msg id ", msgId, " error ")
		c.ExitBuffChan <- true
		return errors.New("conn Write error")
	}

	return nil
}

// NewConntion 创建连接对象，并将服务器注册的路由绑定到该连接。
func NewConntion(
	conn *net.TCPConn,
	connID string,
	msgHandler zitface.IMsgHandle,
	maxPacketSize uint32) zitface.IConnection {
	c := &Connection{
		Conn:          conn,
		ConnID:        connID,
		isClosed:      false,
		ExitBuffChan:  make(chan bool, 1),
		MaxPacketSize: maxPacketSize,
		MsgHandler:    msgHandler,
	}

	return c
}
