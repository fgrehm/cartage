package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/fgrehm/cartage/internal/handler"
	"github.com/fgrehm/cartage/internal/protocol"
)

type mockHandler struct {
	action string
	fn     func(ctx context.Context, payload json.RawMessage) (*protocol.Response, error)
}

func (m *mockHandler) Action() string { return m.action }

func (m *mockHandler) Handle(ctx context.Context, payload json.RawMessage) (*protocol.Response, error) {
	return m.fn(ctx, payload)
}

func sendRequest(t *testing.T, sockPath string, req protocol.Request) protocol.Response {
	t.Helper()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	_, _ = fmt.Fprintf(conn, "%s\n", data)

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	var resp protocol.Response
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

func startServer(t *testing.T, registry *handler.Registry) (string, context.CancelFunc) {
	t.Helper()

	sockPath := filepath.Join(t.TempDir(), "test.sock")
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, sockPath, registry, false)
	}()

	// Wait for socket to be ready
	for i := range 50 {
		conn, err := net.Dial("unix", sockPath)
		if err == nil {
			_ = conn.Close()
			break
		}
		if i == 49 {
			cancel()
			t.Fatal("server didn't start")
		}
		time.Sleep(10 * time.Millisecond)
	}

	return sockPath, cancel
}

func TestServerDispatchSuccess(t *testing.T) {
	reg := handler.NewRegistry()
	reg.Register(&mockHandler{
		action: "test",
		fn: func(_ context.Context, _ json.RawMessage) (*protocol.Response, error) {
			return protocol.OkResponse("hello"), nil
		},
	})

	sockPath, cancel := startServer(t, reg)
	defer cancel()

	resp := sendRequest(t, sockPath, protocol.Request{
		Version: 1,
		Action:  "test",
		Payload: json.RawMessage(`{}`),
	})

	if resp.Status != "ok" {
		t.Errorf("status: want ok, got %s", resp.Status)
	}
}

func TestServerUnknownAction(t *testing.T) {
	reg := handler.NewRegistry()

	sockPath, cancel := startServer(t, reg)
	defer cancel()

	resp := sendRequest(t, sockPath, protocol.Request{
		Version: 1,
		Action:  "nope",
	})

	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestServerInvalidJSON(t *testing.T) {
	reg := handler.NewRegistry()

	sockPath, cancel := startServer(t, reg)
	defer cancel()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()

	_, _ = fmt.Fprintf(conn, "{invalid json\n")

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var resp protocol.Response
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
}

func TestServerGracefulShutdown(t *testing.T) {
	reg := handler.NewRegistry()

	sockPath, cancel := startServer(t, reg)

	// Cancel should stop the server
	cancel()

	// Give it a moment to shut down
	time.Sleep(100 * time.Millisecond)

	// New connections should fail
	_, err := net.Dial("unix", sockPath)
	if err == nil {
		t.Error("expected connection to fail after shutdown")
	}
}
