package secret

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestHandlerAction(t *testing.T) {
	h := &Handler{}
	if h.Action() != "secret" {
		t.Errorf("action: want secret, got %s", h.Action())
	}
}

func TestHandlerInvalidPayload(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{invalid`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestHandlerUnknownOp(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"badop"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestHandlerGetMissingService(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"get","user":"alice"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestHandlerGetMissingUser(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"get","service":"myapp"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestHandlerGetSuccess(t *testing.T) {
	keyring.MockInit()
	if err := keyring.Set("myapp", "alice", "hunter2"); err != nil {
		t.Fatalf("failed to seed keyring: %v", err)
	}

	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"get","service":"myapp","user":"alice"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("status: want ok, got %s (%s)", resp.Status, resp.Error)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("failed to marshal response data: %v", err)
	}
	var r Result
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatalf("failed to decode result: %v", err)
	}
	if r.Secret != "hunter2" {
		t.Errorf("secret: want hunter2, got %s", r.Secret)
	}
}

func TestHandlerGetNotFound(t *testing.T) {
	keyring.MockInit()

	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"get","service":"myapp","user":"nobody"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestHandlerList(t *testing.T) {
	if !keychainAvailable() {
		t.Skip("host keychain unavailable; skipping list test")
	}

	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"list"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("status: want ok, got %s (%s)", resp.Status, resp.Error)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("failed to marshal response data: %v", err)
	}
	var r ListResult
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatalf("failed to decode list result: %v", err)
	}
	if r.Secrets == nil {
		t.Error("secrets should not be nil")
	}
}

func TestHandlerSetSuccess(t *testing.T) {
	keyring.MockInit()

	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"set","service":"myapp","user":"alice","secret":"hunter2"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("status: want ok, got %s (%s)", resp.Status, resp.Error)
	}

	got, err := keyring.Get("myapp", "alice")
	if err != nil {
		t.Fatalf("failed to read back secret: %v", err)
	}
	if got != "hunter2" {
		t.Errorf("secret: want hunter2, got %s", got)
	}
}

func TestHandlerSetMissingService(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"set","user":"alice","secret":"hunter2"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestHandlerSetMissingUser(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"set","service":"myapp","secret":"hunter2"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestHandlerSetMissingSecret(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"set","service":"myapp","user":"alice"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}
