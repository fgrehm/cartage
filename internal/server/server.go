package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fgrehm/cartage/internal/handler"
	"github.com/fgrehm/cartage/internal/protocol"
)

// Run starts the daemon server, listening on the given socket path.
// It dispatches requests to the registry and shuts down gracefully on context cancellation.
func Run(ctx context.Context, socketPath string, registry *handler.Registry, verbose bool) error {
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	if socketPath == "" {
		socketPath = DefaultSocketPath()
	}

	slog.Info("starting daemon", "socket", socketPath)

	// Remove old socket file if it exists
	if _, err := os.Stat(socketPath); err == nil {
		slog.Warn("removing existing socket file", "path", socketPath)
		if err := os.Remove(socketPath); err != nil {
			return fmt.Errorf("failed to remove old socket: %w", err)
		}
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to bind socket %s: %w", socketPath, err)
	}

	if err := os.Chmod(socketPath, 0600); err != nil {
		slog.Warn("failed to set socket permissions", "error", err)
	}

	slog.Info("daemon ready, listening for connections")

	var wg sync.WaitGroup

	// Shutdown goroutine: when ctx is cancelled, close listener and drain connections
	go func() {
		<-ctx.Done()
		slog.Info("shutting down")
		_ = listener.Close()

		// Wait for in-flight connections with a timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			slog.Info("all connections drained")
		case <-time.After(5 * time.Second):
			slog.Warn("shutdown timeout, forcing exit")
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if we're shutting down
			if ctx.Err() != nil {
				break
			}
			slog.Error("accept failed", "error", err)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			handleClient(ctx, conn, registry)
		}()
	}

	// Wait for remaining connections
	wg.Wait()
	return nil
}

// DefaultSocketPath returns the default socket path based on XDG_RUNTIME_DIR.
func DefaultSocketPath() string {
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		return fmt.Sprintf("%s/cartage.sock", xdg)
	}
	return "/tmp/cartage.sock"
}

func handleClient(ctx context.Context, conn net.Conn, registry *handler.Registry) {
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				slog.Error("failed to read line", "error", err)
			}
			break
		}

		slog.Debug("received request", "data", strings.TrimSpace(line))

		var req protocol.Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			slog.Error("invalid JSON", "error", err)
			resp := protocol.ErrorResponse(fmt.Sprintf("invalid JSON: %v", err))
			sendResponse(conn, resp)
			continue
		}

		resp := registry.Dispatch(ctx, &req)
		sendResponse(conn, resp)
	}
}

func sendResponse(conn net.Conn, resp *protocol.Response) {
	jsonData, err := json.Marshal(resp)
	if err != nil {
		slog.Error("failed to serialize response", "error", err)
		return
	}

	slog.Debug("sending response", "data", string(jsonData))

	if _, err := fmt.Fprintf(conn, "%s\n", jsonData); err != nil {
		slog.Error("failed to send response", "error", err)
	}
}
