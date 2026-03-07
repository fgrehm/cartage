package compat

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fgrehm/cartage/internal/client"
	"github.com/fgrehm/cartage/internal/clipboard"
	"github.com/fgrehm/cartage/internal/protocol"
)

// HandlePbpaste handles pbpaste compatibility mode.
// It sends a clipboard read request and prints text to stdout.
// Exits with code 1 if the clipboard contains an image.
func HandlePbpaste(_ []string) {
	payload := clipboard.Payload{Op: clipboard.OpRead}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Send(protocol.Request{
		Version: protocol.CurrentVersion,
		Action:  "clipboard",
		Payload: payloadJSON,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result, err := clipboard.ParseResult(resp.Data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	stat, err := os.Stdout.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to stat stdout: %v\n", err)
		os.Exit(1)
	}
	isTerminal := stat.Mode()&os.ModeCharDevice != 0

	if err := writePbpasteResult(result, isTerminal, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// writePbpasteResult writes the clipboard result to w.
// For text content, it writes the text directly.
// For image content, it writes decoded bytes when not a terminal,
// or returns an error when stdout is a terminal.
func writePbpasteResult(result clipboard.Result, isTerminal bool, w io.Writer) error {
	if result.ContentType == clipboard.ContentImage {
		if isTerminal {
			return fmt.Errorf("clipboard contains an image, not text (use pbpaste > file to save)")
		}
		data, err := base64.StdEncoding.DecodeString(result.ImageData)
		if err != nil {
			return fmt.Errorf("failed to decode image data: %w", err)
		}
		_, err = w.Write(data)
		return err
	}

	_, err := fmt.Fprint(w, result.Text)
	return err
}
