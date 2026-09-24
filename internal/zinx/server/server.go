package server

import (
	"log/slog"
	"net"
	"strconv"
	"sync/atomic"
	"zinx-learn/internal/config"
	"zinx-learn/internal/zinx/routing"
	"zinx-learn/internal/zinx/transport"
)

// Server 负责监听 TCP 连接并分发客户端请求。
type Server struct {
	name              string
	network           string
	host              string
	port              int
	dispatcher        *routing.Dispatcher
	maxPacketSize     uint32
	maxConnections    int
	activeConnections atomic.Int64
}

// Start 异步启动 TCP 监听。
func (s *Server) Start() {
	slog.Info("服务器正在启动", "host", s.host, "port", s.port)

	go func() {
		s.dispatcher.Start()

		address := net.JoinHostPort(s.host, strconv.Itoa(s.port))
		listener, err := net.Listen(s.network, address)
		if err != nil {
			slog.Error("监听失败", "network", s.network, "address", address, "error", err)
			return
		}
		slog.Info("服务器已启动", "name", s.name, "address", listener.Addr())

		var connectionID uint32
		for {
			conn, err := listener.Accept()
			if err != nil {
				slog.Error("接受连接失败", "error", err)
				continue
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
			go func() {
				defer s.activeConnections.Add(-1)
				connection.Start()
			}()
		}
	}()
}

func (s *Server) acquireConnection() bool {
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

// Serve 启动服务器并阻塞调用方。
func (s *Server) Serve() {
	s.Start()
	select {}
}

// Stop 停止网络服务。
func (s *Server) Stop() {
	slog.Info("服务器正在停止", "name", s.name)
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
