package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"zinx-learn/internal/config"
	"zinx-learn/internal/zinx/routing"
	"zinx-learn/internal/zinx/transport"
)

// Server 负责监听 TCP 连接，并协调连接与工作池的生命周期。
type Server struct {
	name              string
	network           string
	host              string
	port              int
	dispatcher        *routing.Dispatcher
	maxPacketSize     uint32
	maxConnections    int
	activeConnections atomic.Int64
	running           atomic.Bool
}

func (s *Server) acquireConnection() bool {
	// 先原子预占一个名额，避免多个连接并发接入时共同越过上限检查。
	for {
		active := s.activeConnections.Load()
		if active >= int64(s.maxConnections) {
			return false
		}
		if s.activeConnections.CompareAndSwap(active, active+1) {
			return true
		}
	}
}

// Serve 启动服务器，并在 ctx 取消后等待连接和工作协程退出。
// 同一个 Server 同一时间只能运行一次；net.Listen 同步执行，使端口占用等启动错误
// 能直接返回给调用方。
func (s *Server) Serve(parent context.Context) error {
	if err := parent.Err(); err != nil {
		return err
	}
	if !s.running.CompareAndSwap(false, true) {
		return errors.New("服务器已在运行")
	}
	defer s.running.Store(false)

	address := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	listener, err := net.Listen(s.network, address)
	if err != nil {
		return fmt.Errorf("监听 %s 失败: %w", address, err)
	}
	slog.Info("服务器已启动", "name", s.name, "address", listener.Addr())
	return s.serve(parent, listener)
}

func (s *Server) serve(parent context.Context, listener net.Listener) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	// dispatcher 与所有连接共享同一个服务级 ctx，停服时由一次取消统一通知。
	dispatcherDone := make(chan struct{})
	go func() {
		defer close(dispatcherDone)
		s.dispatcher.Run(ctx)
	}()

	// Accept 不接收 context；关闭 listener 才能让阻塞中的 Accept 立即返回。
	stopListenerClose := context.AfterFunc(ctx, func() { _ = listener.Close() })
	defer stopListenerClose()

	var connections sync.WaitGroup
	var serveErr error
	// Accept 循环是 connectionID 的唯一写入者，因此这里不需要原子操作或锁。
	var connectionID uint32
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() == nil {
				serveErr = fmt.Errorf("接受连接失败: %w", err)
			}
			break
		}

		if !s.acquireConnection() {
			slog.Warn("连接数已达上限", "max_connections", s.maxConnections, "remote_addr", conn.RemoteAddr())
			if err := conn.Close(); err != nil {
				slog.Debug("拒绝连接时关闭套接字失败", "error", err)
			}
			continue
		}

		connection := transport.NewConnection(conn, connectionID, s.dispatcher, s.maxPacketSize)
		connectionID++
		connections.Add(1)
		go func() {
			defer connections.Done()
			// 连接彻底退出后才归还名额，保证 MaxConn 表示真实存活的连接数。
			defer s.activeConnections.Add(-1)
			connection.Serve(ctx)
		}()
	}

	// 关闭顺序：停止接入和连接 -> 等待连接退出 -> 等待 dispatcher 的 worker 退出。
	cancel()
	_ = listener.Close()
	connections.Wait()
	<-dispatcherDone

	if serveErr != nil {
		return serveErr
	}
	return parent.Err()
}

// AddHandler 为消息 ID 注册处理器。
func (s *Server) AddHandler(messageID uint32, handler routing.Handler) error {
	return s.dispatcher.Register(messageID, handler)
}

// New 根据配置创建服务器。
func New(cfg config.ServerConfig, name string) *Server {
	return &Server{
		name:           name,
		network:        cfg.IPVersion,
		host:           cfg.Host,
		port:           cfg.Port,
		dispatcher:     routing.NewDispatcher(cfg.WorkerPoolSize, cfg.MaxWorkerTaskLen),
		maxPacketSize:  cfg.MaxPacketSize,
		maxConnections: cfg.MaxConn,
	}
}
