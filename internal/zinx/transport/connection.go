package transport

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"zinx-learn/internal/zinx/protocol"
	"zinx-learn/internal/zinx/routing"
)

// Connection 封装一个客户端 TCP 连接。
type Connection struct {
	conn          net.Conn
	id            uint32
	done          chan struct{}
	closeOnce     sync.Once
	closeErr      error
	maxPacketSize uint32
	dispatcher    *routing.Dispatcher
	msgChan       chan []byte
}

// 检验 Connection 是否实现了 Responder 接口
var _ routing.Responder = (*Connection)(nil)

// NewConnection 创建一个连接。
func NewConnection(
	conn net.Conn,
	id uint32,
	dispatcher *routing.Dispatcher,
	maxPacketSize uint32,
) *Connection {
	return &Connection{
		conn:          conn,
		id:            id,
		done:          make(chan struct{}),
		maxPacketSize: maxPacketSize,
		dispatcher:    dispatcher,
		msgChan:       make(chan []byte),
	}
}

func (cn *Connection) writeLoop() {
	slog.Debug("连接写协程已启动", "remote_addr", cn.RemoteAddr())
	defer slog.Debug("连接写协程已退出", "remote_addr", cn.RemoteAddr())
	defer func() { _ = cn.Close() }()

	for {
		select {
		case data := <-cn.msgChan:
			if _, err := cn.conn.Write(data); err != nil {
				if !errors.Is(err, net.ErrClosed) {
					slog.Error("发送数据失败", "connection_id", cn.id, "error", err)
				}
				return
			}
		case <-cn.done:
			return
		}
	}
}

// SendMessage 将消息封包后发送给客户端。
func (cn *Connection) SendMessage(messageID uint32, data []byte) error {
	packer := protocol.NewDataPack(cn.maxPacketSize)
	message, err := packer.Pack(protocol.NewMessage(messageID, data))
	if err != nil {
		return fmt.Errorf("封装消息 %d 失败: %w", messageID, err)
	}

	select {
	case cn.msgChan <- message:
		return nil
	case <-cn.done:
		return net.ErrClosed
	}
}

// ID 返回连接标识。
func (cn *Connection) ID() uint32 {
	return cn.id
}

// RemoteAddr 返回客户端地址。
func (cn *Connection) RemoteAddr() net.Addr {
	return cn.conn.RemoteAddr()
}

// Close 关闭底层套接字并通知读写协程退出。
func (cn *Connection) Close() error {
	cn.closeOnce.Do(func() {
		close(cn.done)
		cn.closeErr = cn.conn.Close()
	})
	return cn.closeErr
}

func (cn *Connection) readLoop() {
	slog.Debug("连接读协程已启动", "remote_addr", cn.RemoteAddr())
	defer slog.Debug("连接读协程已退出", "remote_addr", cn.RemoteAddr())
	defer func() { _ = cn.Close() }()

	packer := protocol.NewDataPack(cn.maxPacketSize)
	for {
		header := make([]byte, packer.HeaderLen())
		if _, err := io.ReadFull(cn.conn, header); err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
				slog.Error("读取消息包头失败", "connection_id", cn.id, "error", err)
			}
			return
		}

		messageID, dataLen, err := packer.UnpackHeader(header)
		if err != nil {
			slog.Error("解析消息包头失败", "connection_id", cn.id, "error", err)
			return
		}

		data := make([]byte, dataLen)
		if dataLen > 0 {
			if _, err := io.ReadFull(cn.conn, data); err != nil {
				if !errors.Is(err, net.ErrClosed) {
					slog.Error("读取消息体失败", "connection_id", cn.id, "error", err)
				}
				return
			}
		}

		cn.dispatcher.Submit(routing.NewRequest(cn, cn.id, messageID, data))
	}
}

// Start 启动连接读写协程，并等待它们退出。
func (cn *Connection) Start() {
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		cn.readLoop()
	}()
	go func() {
		defer workers.Done()
		cn.writeLoop()
	}()
	workers.Wait()
}
