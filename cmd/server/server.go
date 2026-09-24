package main

import (
	"log/slog"
	"os"
	"zinx-learn/internal/config"
	"zinx-learn/internal/zinx/routing"
	"zinx-learn/internal/zinx/server"
)

type PingHandler struct{}

func (h *PingHandler) Handle(request *routing.Request) {
	slog.Info("收到消息", "message_id", request.MessageID(), "data", string(request.Data()))
	if err := request.Responder().SendMessage(1, []byte("ping...ping...ping")); err != nil {
		slog.Error("回复消息失败", "error", err)
	}
}

type HelloHandler struct{}

func (h *HelloHandler) Handle(request *routing.Request) {
	slog.Info("收到消息", "message_id", request.MessageID(), "data", string(request.Data()))
	if err := request.Responder().SendMessage(1, []byte("Hello Zinx Handler v0.8.1")); err != nil {
		slog.Error("回复消息失败", "error", err)
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	srv := server.New(cfg.Server, "zinx v0.8.1")
	if err := srv.AddHandler(0, &PingHandler{}); err != nil {
		slog.Error("注册处理器失败", "error", err)
		os.Exit(1)
	}
	if err := srv.AddHandler(1, &HelloHandler{}); err != nil {
		slog.Error("注册处理器失败", "error", err)
		os.Exit(1)
	}

	srv.Serve()
}
