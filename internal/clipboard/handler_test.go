package clipboard

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHandlerAction(t *testing.T) {
	h := &Handler{}
	if h.Action() != "clipboard" {
		t.Errorf("action: want clipboard, got %s", h.Action())
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

func TestHandlerWriteMissingContent(t *testing.T) {
	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"op":"write"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestHandlerWriteBothTextAndImage(t *testing.T) {
	h := &Handler{}
	text := "hello"
	imgData := "aW1hZ2VkYXRh" // base64 of "imagedata"
	p := Payload{
		Op:        OpWrite,
		Text:      &text,
		ImageData: &imgData,
	}
	raw, _ := json.Marshal(p)
	resp, err := h.Handle(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}
