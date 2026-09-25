package main

import (
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"
	"zinx-learn/internal/config"
	"zinx-learn/internal/zinx/protocol"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	slog.Info("客户端正在启动")

	// 这是演示程序的简单等待，不是生产环境中的服务就绪检测机制。
	time.Sleep(3 * time.Second)

	address := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))
	conn, err := net.Dial(cfg.Server.IPVersion, address)
	if err != nil {
		slog.Error("连接服务器失败", "error", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Debug("关闭客户端连接失败", "error", err)
		}
	}()

	packer := protocol.NewDataPack(cfg.Server.MaxPacketSize)

	// 演示客户端采用同步的“发送一条、读取一条”流程，便于观察完整通信过程。
	for {
		message, err := packer.Pack(protocol.NewMessage(0, []byte("Zinx v0.9.0 Client Test Message")))
		if err != nil {
			slog.Error("封装消息失败", "error", err)
			return
		}
		_, err = conn.Write(message)
		if err != nil {
			slog.Error("发送消息失败", "error", err)
			return
		}

		// TCP 是字节流，需要先读满固定包头，再按其中的长度读满响应体。
		header := make([]byte, packer.HeaderLen())
		_, err = io.ReadFull(conn, header)
		if err != nil {
			slog.Error("读取响应包头失败", "error", err)
			return
		}

		messageID, dataLen, err := packer.UnpackHeader(header)
		if err != nil {
			slog.Error("解析响应包头失败", "error", err)
			return
		}

		data := make([]byte, dataLen)
		if dataLen > 0 {
			_, err := io.ReadFull(conn, data)
			if err != nil {
				slog.Error("读取响应消息体失败", "error", err)
				return
			}
		}
		slog.Info("收到响应", "message_id", messageID, "length", dataLen, "data", string(data))

		time.Sleep(1 * time.Second)
	}
}
