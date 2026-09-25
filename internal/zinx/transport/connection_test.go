package transport

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
	"zinx-learn/internal/zinx/routing"
)

func TestConnectionStopsAfterPeerCloses(t *testing.T) {
	// net.Pipe 提供一对内存连接，可以在不监听真实端口的情况下测试连接生命周期。
	serverConn, clientConn := net.Pipe()
	dispatcher := routing.NewDispatcher(1, 1)
	connection := NewConnection(serverConn, 1, dispatcher, 1024)

	stopped := make(chan struct{})
	go func() {
		connection.Serve(context.Background())
		close(stopped)
	}()

	if err := clientConn.Close(); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("connection did not stop after peer closed")
	}

	if err := connection.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		t.Fatalf("second Close() error = %v", err)
	}
	if err := connection.SendMessage(context.Background(), 1, nil); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("SendMessage() error = %v, want %v", err, net.ErrClosed)
	}
}

func TestConnectionStopsAfterContextCancellation(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() { _ = clientConn.Close() })

	connection := NewConnection(serverConn, 1, routing.NewDispatcher(1, 1), 1024)
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	go func() {
		connection.Serve(ctx)
		close(stopped)
	}()

	cancel()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("connection did not stop after context cancellation")
	}
}

func TestSendMessageHonorsCanceledContext(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() { _ = serverConn.Close() })
	t.Cleanup(func() { _ = clientConn.Close() })

	connection := NewConnection(serverConn, 1, routing.NewDispatcher(1, 1), 1024)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := connection.SendMessage(ctx, 1, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SendMessage() error = %v, want %v", err, context.Canceled)
	}
}
