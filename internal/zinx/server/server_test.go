package server

import "testing"

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
