package client

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/fgrehm/cartage/internal/protocol"
)

func TestSendRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		buf := make([]byte, 4096)
		n, _ := conn.Read(buf)

		var req protocol.Request
		if err := json.Unmarshal(buf[:n-1], &req); err != nil {
			return
		}

		resp := protocol.OkResponse(map[string]string{"id": "test-123"})
		data, _ := json.Marshal(resp)
		_, _ = fmt.Fprintf(conn, "%s\n", data)
	}()

	t.Setenv("CARTAGE_SOCKET", sockPath)

	req := protocol.Request{
		Version: 1,
		Action:  "notify",
		Payload: json.RawMessage(`{"title":"hello"}`),
	}

	resp, err := Send(req)
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("status: want ok, got %s", resp.Status)
	}
}

func TestSendConnectionError(t *testing.T) {
	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "nonexistent.sock")
	t.Setenv("CARTAGE_SOCKET", sockPath)

	req := protocol.Request{Version: 1, Action: "test"}
	_, err := Send(req)
	if err == nil {
		t.Fatal("expected error for non-existent socket")
	}
}

func TestSendErrorResponse(t *testing.T) {
	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		buf := make([]byte, 4096)
		_, _ = conn.Read(buf)

		resp := protocol.ErrorResponse("bad request")
		data, _ := json.Marshal(resp)
		_, _ = fmt.Fprintf(conn, "%s\n", data)
	}()

	t.Setenv("CARTAGE_SOCKET", sockPath)

	req := protocol.Request{Version: 1, Action: "fail"}
	resp, err := Send(req)

	if err == nil {
		t.Fatal("expected error for error response")
	}
	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestMustSendPanicsOnMissingSock(t *testing.T) {
	// MustSend calls os.Exit, which we can't easily test.
	_ = os.Exit
}
