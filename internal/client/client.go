package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/fgrehm/cartage/internal/protocol"
)

// Send sends a request to the daemon and returns the response.
func Send(req protocol.Request) (protocol.Response, error) {
	socketPath := FindSocketPath()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return protocol.Response{}, fmt.Errorf(
			"failed to connect to daemon at %s: %w\n\nIs the daemon running? Try: cartage serve",
			socketPath, err)
	}
	defer func() { _ = conn.Close() }()

	jsonData, err := json.Marshal(req)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("failed to serialize request: %w", err)
	}

	if _, err := fmt.Fprintf(conn, "%s\n", jsonData); err != nil {
		return protocol.Response{}, fmt.Errorf("failed to send request: %w", err)
	}

	reader := bufio.NewReader(conn)
	responseLine, err := reader.ReadString('\n')
	if err != nil {
		return protocol.Response{}, fmt.Errorf("failed to read response: %w", err)
	}

	var resp protocol.Response
	if err := json.Unmarshal([]byte(responseLine), &resp); err != nil {
		return protocol.Response{}, fmt.Errorf(
			"invalid response from daemon: %w\nResponse: %s",
			err, responseLine)
	}

	if resp.Status == "error" {
		return resp, fmt.Errorf("daemon returned error: %s", resp.Error)
	}

	return resp, nil
}

// MustSend is a helper that calls Send and exits on error.
func MustSend(req protocol.Request) protocol.Response {
	resp, err := Send(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return resp
}
