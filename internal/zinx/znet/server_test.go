package znet

import (
	"testing"
	"zinx-learn/internal/config"
)

func TestNewServer(t *testing.T) {
	const name = "[zinx V0.3]"
	cfg := &config.Config{
		Server: config.ServerConfig{
			IPVersion: "tcp4",
			Host:      "127.0.0.1",
			Port:      7777,
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
}

func TestServerAddRouter(t *testing.T) {
	server := &Server{}
	router := &BaseRouter{}

	server.AddRouter(router)

	if server.Router != router {
		t.Fatal("AddRouter() did not register the provided router")
	}
}
