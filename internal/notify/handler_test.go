package notify

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHandlerAction(t *testing.T) {
	h := &Handler{}
	if h.Action() != "notify" {
		t.Errorf("action: want notify, got %s", h.Action())
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

func TestHandlerMissingTitle(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}
