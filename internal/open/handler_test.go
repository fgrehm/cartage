package open

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestHandlerXdgOpenStderrIncludedInError(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "xdg-open")
	err := os.WriteFile(script, []byte("#!/bin/sh\necho 'could not connect to display' >&2\nexit 3\n"), 0o755)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))

	h := &Handler{}
	resp, err := h.Handle(context.Background(), json.RawMessage(`{"uri":"https://example.com"}`))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "could not connect to display") {
		t.Errorf("error %q should contain stderr output", resp.Error)
	}
}
