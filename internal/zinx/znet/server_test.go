package znet

import (
	"testing"
	"zinx-learn/internal/config"
)

func TestNewServer(t *testing.T) {
	const name = "[zinx V0.5]"
	cfg := &config.Config{
		Server: config.ServerConfig{
			IPVersion:     "tcp4",
			Host:          "127.0.0.1",
			Port:          7777,
			MaxPacketSize: 4096,
		},
	}

	server, ok := NewServer(cfg, name).(*Server)
	if !ok {
		t.Fatal("NewServer() did not return *Server")
	}

	if server.Name != name {
		t.Errorf("Name = %q, want %q", server.Name, name)
	}
	if server.IPVersion != "tcp4" {
		t.Errorf("IPVersion = %q, want %q", server.IPVersion, "tcp4")
	}
	if server.IP != "127.0.0.1" {
		t.Errorf("IP = %q, want %q", server.IP, "127.0.0.1")
	}
	if server.Port != 7777 {
		t.Errorf("Port = %d, want %d", server.Port, 7777)
	}
	if server.MaxPacketSize != 4096 {
		t.Errorf("MaxPacketSize = %d, want %d", server.MaxPacketSize, 4096)
	}
	if server.msgHandler == nil {
		t.Fatal("NewServer() did not initialize the message handler")
	}
}

func TestServerAddRouter(t *testing.T) {
	server := &Server{msgHandler: NewMsgHandler()}
	router := &BaseRouter{}

	server.AddRouter(1, router)

	handler := server.msgHandler.(*MsgHandler)
	if handler.Apis[1] != router {
		t.Fatal("AddRouter() did not register the provided router")
	}
}
