package znet

import (
	"bytes"
	"testing"
)

func TestRequest(t *testing.T) {
	conn := &Connection{ConnID: "conn-1"}
	data := []byte("ping")
	request := &Request{
		conn: conn,
		data: data,
	}

	if request.GetConnection() != conn {
		t.Fatal("GetConnection() did not return the request connection")
	}
	if got := request.GetData(); !bytes.Equal(got, data) {
		t.Fatalf("GetData() = %q, want %q", got, data)
	}
}

func TestBaseRouterDefaultHandlers(t *testing.T) {
	router := &BaseRouter{}
	request := &Request{}

	// 三个默认处理方法都应为空操作，调用时不应发生 panic。
	router.PreHandle(request)
	router.Handle(request)
	router.PostHandle(request)
}
