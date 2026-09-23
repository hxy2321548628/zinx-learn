package main

import (
	"fmt"
	"log"
	"zinx-learn/internal/config"
	"zinx-learn/internal/zinx/zitface"
	"zinx-learn/internal/zinx/znet"
)

// PingRouter 演示一个包含前置、核心和后置处理阶段的自定义路由。
type PingRouter struct {
	// 嵌入 BaseRouter 后，只需重写业务需要的处理阶段。
	znet.BaseRouter
}

// PreHandle 在核心处理前向客户端发送提示消息。
func (this *PingRouter) PreHandle(request zitface.IRequest) {
	fmt.Println("Call Router PreHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("before ping ....\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

// Handle 执行 ping 请求的核心响应逻辑。
func (this *PingRouter) Handle(request zitface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

// PostHandle 在核心处理结束后向客户端发送收尾消息。
func (this *PingRouter) PostHandle(request zitface.IRequest) {
	fmt.Println("Call Router PostHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("After ping .....\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

// Server 模块的测试函数
func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	//1 创建一个server 句柄 s
	s := znet.NewServer(cfg, "[zinx V0.3]")
	log.Printf("服务监听地址：%s:%d", cfg.Server.Host, cfg.Server.Port)

	// 将自定义路由注册到服务器，后续建立的连接都会使用该路由处理请求。
	s.AddRouter(&PingRouter{})

	//2 开启服务
	s.Serve()
}
