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

	//3秒之后发起测试请求，给服务端开启服务的机会
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

	for {
		message, err := packer.Pack(protocol.NewMessage(1, []byte("Zinx v0.8.1 Client Test Message")))
		if err != nil {
			slog.Error("封装消息失败", "error", err)
			return
		}
		_, err = conn.Write(message)
		if err != nil {
			slog.Error("发送消息失败", "error", err)
			return
		}

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
