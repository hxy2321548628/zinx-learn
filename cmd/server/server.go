package main

import (
	"log"
	"zinx-learn/internal/config"
	"zinx-learn/internal/zinx/znet"
)

// Server 模块的测试函数
func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	//1 创建一个server 句柄 s
	s := znet.NewServer(cfg, "[zinx V0.1]")
	log.Printf("服务监听地址：%s:%d", cfg.Server.Host, cfg.Server.Port)

	//2 开启服务
	s.Serve()
}
