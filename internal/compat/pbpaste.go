package compat

import (
	"encoding/json"
	"fmt"
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

	if result.ContentType == clipboard.ContentImage {
		fmt.Fprintln(os.Stderr, "Error: clipboard contains an image, not text")
		os.Exit(1)
	}

	fmt.Print(result.Text)
	os.Exit(0)
}
