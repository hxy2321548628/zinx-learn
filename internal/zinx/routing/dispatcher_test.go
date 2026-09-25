package routing

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDispatcherRegisterAndDispatch(t *testing.T) {
	t.Parallel()

	dispatcher := NewDispatcher(1, 1)
	called := make(chan struct{})
	handler := HandlerFunc(func(ctx context.Context, request *Request) error {
		if request.MessageID() == 9 {
			close(called)
		}
		return nil
	})

	if err := dispatcher.Register(9, handler); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan struct{})
	go func() {
		dispatcher.Run(ctx)
		close(runDone)
	}()

	if err := dispatcher.Submit(ctx, NewRequest(nil, 1, 9, nil)); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("registered handler was not called")
	}

	cancel()
	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after context cancellation")
	}
}

func TestDispatcherRejectsInvalidRegistration(t *testing.T) {
	t.Parallel()

	dispatcher := NewDispatcher(1, 1)
	if err := dispatcher.Register(1, nil); err == nil {
		t.Fatal("Register() accepted a nil handler")
	}
	handler := HandlerFunc(func(context.Context, *Request) error { return nil })
	if err := dispatcher.Register(1, handler); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if err := dispatcher.Register(1, handler); err == nil {
		t.Fatal("Register() accepted a duplicate message ID")
	}
}

func TestDispatcherSubmitHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	dispatcher := NewDispatcher(1, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := dispatcher.Submit(ctx, NewRequest(nil, 1, 1, nil))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Submit() error = %v, want %v", err, context.Canceled)
	}
}
