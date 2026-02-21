package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fgrehm/cartage/internal/protocol"
)

// mockHandler is a test double implementing Handler.
type mockHandler struct {
	action string
	fn     func(ctx context.Context, payload json.RawMessage) (*protocol.Response, error)
}

func (m *mockHandler) Action() string { return m.action }

func (m *mockHandler) Handle(ctx context.Context, payload json.RawMessage) (*protocol.Response, error) {
	return m.fn(ctx, payload)
}

func TestRegistryDispatchSuccess(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockHandler{
		action: "test",
		fn: func(_ context.Context, _ json.RawMessage) (*protocol.Response, error) {
			return protocol.OkResponse("done"), nil
		},
	})

	resp := reg.Dispatch(context.Background(), &protocol.Request{
		Version: 1,
		Action:  "test",
	})

	if resp.Status != "ok" {
		t.Errorf("status: want ok, got %s", resp.Status)
	}
}

func TestRegistryDispatchUnknownAction(t *testing.T) {
	reg := NewRegistry()

	resp := reg.Dispatch(context.Background(), &protocol.Request{
		Version: 1,
		Action:  "nope",
	})

	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
	if resp.Error == "" {
		t.Error("error message should not be empty")
	}
}

func TestRegistryDispatchBadVersion(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockHandler{
		action: "test",
		fn: func(_ context.Context, _ json.RawMessage) (*protocol.Response, error) {
			return protocol.OkResponse(nil), nil
		},
	})

	resp := reg.Dispatch(context.Background(), &protocol.Request{
		Version: 99,
		Action:  "test",
	})

	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestRegistryDispatchMissingVersionTreatedAsV1(t *testing.T) {
	reg := NewRegistry()
	called := false
	reg.Register(&mockHandler{
		action: "test",
		fn: func(_ context.Context, _ json.RawMessage) (*protocol.Response, error) {
			called = true
			return protocol.OkResponse(nil), nil
		},
	})

	// Version 0 (zero value, i.e., missing from JSON) should be treated as v1
	resp := reg.Dispatch(context.Background(), &protocol.Request{
		Version: 0,
		Action:  "test",
	})

	if resp.Status != "ok" {
		t.Errorf("status: want ok, got %s", resp.Status)
	}
	if !called {
		t.Error("handler should have been called")
	}
}

func TestRegistryDispatchHandlerError(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockHandler{
		action: "fail",
		fn: func(_ context.Context, _ json.RawMessage) (*protocol.Response, error) {
			return nil, fmt.Errorf("boom")
		},
	})

	resp := reg.Dispatch(context.Background(), &protocol.Request{
		Version: 1,
		Action:  "fail",
	})

	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
	if resp.Error != "boom" {
		t.Errorf("error: want 'boom', got %s", resp.Error)
	}
}

func TestRegistryDispatchPassesPayload(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockHandler{
		action: "echo",
		fn: func(_ context.Context, payload json.RawMessage) (*protocol.Response, error) {
			return protocol.OkResponse(string(payload)), nil
		},
	})

	resp := reg.Dispatch(context.Background(), &protocol.Request{
		Version: 1,
		Action:  "echo",
		Payload: json.RawMessage(`{"msg":"hi"}`),
	})

	if resp.Status != "ok" {
		t.Errorf("status: want ok, got %s", resp.Status)
	}
}
