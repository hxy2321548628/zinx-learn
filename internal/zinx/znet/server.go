package znet

import (
	"fmt"
	"net"
	"zinx-learn/internal/config"
	"zinx-learn/internal/zinx/zitface"

	"github.com/google/uuid"
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

			connhand := NewConntion(conn.(*net.TCPConn), uuid.NewString(), CallBackToClient)
			go connhand.Start()

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
func NewServer(config *config.Config, name string) zitface.IServer {
	s := &Server{
		Name:      name,
		IPVersion: config.Server.IPVersion,
		IP:        config.Server.Host,
		Port:      config.Server.Port,
	}
	return s
}
