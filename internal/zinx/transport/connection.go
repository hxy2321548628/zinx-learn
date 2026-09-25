package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"zinx-learn/internal/zinx/protocol"
	"zinx-learn/internal/zinx/routing"
)

// Connection 封装一个客户端 TCP 连接及其读写生命周期。
//
// 每个连接只启动一个读协程和一个写协程。所有响应都先进入 msgChan，再由唯一的
// 写协程写入套接字，避免多个 Handler 并发写入时破坏消息帧边界。
type Connection struct {
	conn net.Conn
	id   uint32
	// done 由 Close 关闭，用作连接级的退出广播；closeOnce 保证该操作幂等。
	done          chan struct{}
	closeOnce     sync.Once
	closeErr      error
	maxPacketSize uint32
	dispatcher    *routing.Dispatcher
	// 无缓冲通道把网络写入的背压传递给发送方，避免响应在内存中无限堆积。
	msgChan chan []byte
}

// 编译期检查 Connection 是否实现了 routing.Responder。
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

func (cn *Connection) writeLoop(ctx context.Context) {
	slog.Debug("连接写协程已启动", "remote_addr", cn.RemoteAddr())
	defer slog.Debug("连接写协程已退出", "remote_addr", cn.RemoteAddr())

	for {
		select {
		case data := <-cn.msgChan:
			// 只有该协程调用 Write，因此每个已封装消息都会完整、顺序地写入连接。
			if _, err := cn.conn.Write(data); err != nil {
				if ctx.Err() == nil && !errors.Is(err, net.ErrClosed) {
					slog.Error("发送数据失败", "connection_id", cn.id, "error", err)
				}
				return
			}
		case <-cn.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// SendMessage 将消息封包后交给连接的唯一写协程。
// ctx 让发送方在连接拥塞或服务器关闭时停止等待。
func (cn *Connection) SendMessage(ctx context.Context, messageID uint32, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

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
	case <-ctx.Done():
		return ctx.Err()
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
// 它可以被读协程、写协程和服务器关闭流程重复调用，实际关闭只会执行一次。
func (cn *Connection) Close() error {
	cn.closeOnce.Do(func() {
		close(cn.done)
		cn.closeErr = cn.conn.Close()
	})
	return cn.closeErr
}

func (cn *Connection) readLoop(ctx context.Context) {
	slog.Debug("连接读协程已启动", "remote_addr", cn.RemoteAddr())
	defer slog.Debug("连接读协程已退出", "remote_addr", cn.RemoteAddr())

	packer := protocol.NewDataPack(cn.maxPacketSize)
	for {
		// TCP 没有消息边界：先精确读取固定包头，再按包头声明的长度读取消息体。
		header := make([]byte, packer.HeaderLen())
		if _, err := io.ReadFull(cn.conn, header); err != nil {
			if ctx.Err() == nil && !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
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
				if ctx.Err() == nil && !errors.Is(err, net.ErrClosed) {
					slog.Error("读取消息体失败", "connection_id", cn.id, "error", err)
				}
				return
			}
		}

		if err := cn.dispatcher.Submit(ctx, routing.NewRequest(cn, cn.id, messageID, data)); err != nil {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				slog.Error("提交请求失败", "connection_id", cn.id, "error", err)
			}
			return
		}
	}
}

// Serve 启动连接读写协程，并等待它们退出。
//
// context 的取消本身不能打断阻塞中的 net.Conn.Read/Write，因此 AfterFunc 会关闭
// 套接字来唤醒系统调用。任一读写协程先退出时，也会取消同级协程并关闭连接。
func (cn *Connection) Serve(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	stopClose := context.AfterFunc(ctx, func() { _ = cn.Close() })
	defer stopClose()

	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		defer cancel()
		defer func() { _ = cn.Close() }()
		cn.readLoop(ctx)
	}()
	go func() {
		defer workers.Done()
		defer cancel()
		defer func() { _ = cn.Close() }()
		cn.writeLoop(ctx)
	}()
	workers.Wait()
}
