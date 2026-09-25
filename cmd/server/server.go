package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"zinx-learn/internal/config"
	"zinx-learn/internal/zinx/routing"
	"zinx-learn/internal/zinx/server"
)

type PingHandler struct{}

func (h *PingHandler) Handle(ctx context.Context, request *routing.Request) error {
	slog.Info("收到消息", "message_id", request.MessageID(), "data", string(request.Data()))
	// 将发送错误返回给 dispatcher，由框架统一记录处理失败日志。
	return request.Responder().SendMessage(ctx, 1, []byte("ping...ping...ping"))
}

type HelloHandler struct{}

func (h *HelloHandler) Handle(ctx context.Context, request *routing.Request) error {
	slog.Info("收到消息", "message_id", request.MessageID(), "data", string(request.Data()))
	// Handler 只依赖 Responder，不需要知道底层使用的是 TCP 还是测试替身。
	return request.Responder().SendMessage(ctx, 1, []byte("Hello Zinx Handler v0.9.0"))
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	srv := server.New(cfg.Server, "zinx v0.9.0")
	// 在启动工作池前完成路由注册，避免请求到达时处理器尚未就绪。
	if err := srv.AddHandler(0, &PingHandler{}); err != nil {
		slog.Error("注册处理器失败", "error", err)
		os.Exit(1)
	}
	if err := srv.AddHandler(1, &HelloHandler{}); err != nil {
		slog.Error("注册处理器失败", "error", err)
		os.Exit(1)
	}

	// 信号 context 是整棵并发任务树的根；收到退出信号后，取消会传播到所有连接和 worker。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// context.Canceled 表示信号触发的正常停服，不应作为故障退出。
	if err := srv.Serve(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("服务器退出", "error", err)
		os.Exit(1)
	}
	slog.Info("服务器已停止")
}
