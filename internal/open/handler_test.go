package open

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHandlerAction(t *testing.T) {
	h := &Handler{}
	if h.Action() != "open" {
		t.Errorf("action: want open, got %s", h.Action())
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

func TestHandlerEmptyURI(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"uri":""}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestHandlerMissingURI(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}
