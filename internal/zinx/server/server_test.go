package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"
	"zinx-learn/internal/zinx/routing"
)

func TestServerConnectionLimit(t *testing.T) {
	t.Parallel()

	server := &Server{maxConnections: 1}
	if !server.acquireConnection() {
		t.Fatal("first connection was rejected")
	}
	if server.acquireConnection() {
		t.Fatal("connection over the limit was accepted")
	}

	server.activeConnections.Add(-1)
	if !server.acquireConnection() {
		t.Fatal("connection slot was not released")
	}
}

func TestServeStopsAfterContextCancellation(t *testing.T) {
	t.Parallel()

	server := &Server{
		dispatcher:     routing.NewDispatcher(1, 1),
		maxConnections: 1,
	}
	listener := newBlockingListener()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- server.serve(ctx, listener)
	}()

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("serve() error = %v, want %v", err, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Fatal("serve() did not stop after context cancellation")
	}
}

// blockingListener 模拟阻塞中的 Accept，只有 Close 才会将其唤醒。
// 这用于验证取消 context 后，服务器确实会主动关闭 listener，而不是泄漏 Accept 协程。
type blockingListener struct {
	closed    chan struct{}
	closeOnce sync.Once
}

func newBlockingListener() *blockingListener {
	return &blockingListener{closed: make(chan struct{})}
}

func (l *blockingListener) Accept() (net.Conn, error) {
	<-l.closed
	return nil, net.ErrClosed
}

func (l *blockingListener) Close() error {
	l.closeOnce.Do(func() { close(l.closed) })
	return nil
}

func (l *blockingListener) Addr() net.Addr {
	return testAddr("test")
}

type testAddr string

func (a testAddr) Network() string { return string(a) }
func (a testAddr) String() string  { return string(a) }
