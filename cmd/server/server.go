package main

import (
	"fmt"
	"log"
	"zinx-learn/internal/config"
	"zinx-learn/internal/zinx/zitface"
	"zinx-learn/internal/zinx/znet"
)

// ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

// Test Handle
func (this *PingRouter) Handle(request zitface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	//回写数据
	err := request.GetConnection().SendMsg(1, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
	}
}

// HelloZinxRouter Handle
type HelloZinxRouter struct {
	znet.BaseRouter
}

func (this *HelloZinxRouter) Handle(request zitface.IRequest) {
	fmt.Println("Call HelloZinxRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendMsg(1, []byte("Hello Zinx Router V0.5"))
	if err != nil {
		fmt.Println(err)
	}
}

// Server 模块的测试函数
func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	//1 创建一个server 句柄 s
	s := znet.NewServer(cfg.Server, "[zinx V0.5]")
	log.Printf("服务监听地址：%s:%d", cfg.Server.Host, cfg.Server.Port)

	// 将自定义路由注册到服务器，后续建立的连接都会使用该路由处理请求。
	//配置路由
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})

	//2 开启服务
	s.Serve()
}
