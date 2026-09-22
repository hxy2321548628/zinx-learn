package znet

import (
	"fmt"
	"net"
	"zinx-learn/internal/zinx/zitface"
)

// iServer 接口实现，定义一个Server服务类
type Server struct {
	Name      string //服务器的名称
	IPVersion string //tcp4 or other
	IP        string //服务绑定的IP地址
	Port      int    //服务绑定的端口
}

//============== 实现 ziface.IServer 里的全部接口方法 ========

// 开启网络服务
func (this *Server) Start() {
	fmt.Printf("[START] Server listenner at IP: %s, Port %d, is starting\n", this.IP, this.Port)

	//开启一个go去做服务端Linster业务
	go func() {
		// 1. 创建一个 tcp socket
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
		if err != nil {
			fmt.Println("listen", this.IPVersion, "err", err)
			return
		}
		//已经监听成功
		fmt.Println("Start Zinx server  ", this.Name, " succ, now listenning...")

		// 2. 监听连接处理业务
		for {
			// 阻塞等待客户端建立连接请求
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}

			//3.2 TODO Server.Start() 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接

			//3.3 TODO Server.Start() 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的

			//我们这里暂时做一个最大512字节的回显服务
			go func() {
				//不断的循环从客户端获取数据
				for {
					buf := make([]byte, 512)
					count, err := conn.Read(buf)
					if err != nil {
						fmt.Println("recv buf err ", err)
						continue
					}

					fmt.Printf(" receive client msg : %s, count = %d\n", buf, count)

					//回显
					if _, err := conn.Write(buf[:count]); err != nil {
						fmt.Println("write back buf err ", err)
						continue
					}
				}
			}()

		}
	}()

}

// 启动和结束的封装, 对外暴露
func (this *Server) Serve() {

	this.Start()

	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则主Go退出， listenner的go将会退出
	select {}

}

// 暂停网络服务
func (this *Server) Stop() {
	fmt.Println("[STOP] Zinx server , name ", this.Name)

	//TODO  Server.Stop() 将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
}

// 服务器实例化函数
func NewServer(name string) zitface.IServer {
	s := &Server{
		Name:      name,
		IPVersion: "tcp4",
		IP:        "127.0.0.1",
		Port:      7777,
	}
	return s
}
