package transport

import (
	"errors"
	"net"
	"testing"
	"time"
	"zinx-learn/internal/zinx/routing"
)

func TestConnectionStopsAfterPeerCloses(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	dispatcher := routing.NewDispatcher(1, 1)
	connection := NewConnection(serverConn, 1, dispatcher, 1024)

	stopped := make(chan struct{})
	go func() {
		connection.Start()
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
	if err := connection.SendMessage(1, nil); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("SendMessage() error = %v, want %v", err, net.ErrClosed)
	}
}
