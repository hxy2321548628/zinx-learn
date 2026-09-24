package routing

import "testing"

func TestDispatcherRegisterAndDispatch(t *testing.T) {
	t.Parallel()

	dispatcher := NewDispatcher(1, 1)
	called := false
	handler := HandlerFunc(func(request *Request) {
		called = request.MessageID() == 9
	})

	if err := dispatcher.Register(9, handler); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	dispatcher.dispatch(NewRequest(nil, 1, 9, nil))
	if !called {
		t.Fatal("registered handler was not called")
	}
}

func TestDispatcherRejectsInvalidRegistration(t *testing.T) {
	t.Parallel()

	dispatcher := NewDispatcher(1, 1)
	if err := dispatcher.Register(1, nil); err == nil {
		t.Fatal("Register() accepted a nil handler")
	}
	if err := dispatcher.Register(1, HandlerFunc(func(*Request) {})); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if err := dispatcher.Register(1, HandlerFunc(func(*Request) {})); err == nil {
		t.Fatal("Register() accepted a duplicate message ID")
	}
}
