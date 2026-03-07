package compat

import (
	"encoding/base64"
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
		stat, err := os.Stdout.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to stat stdout: %v\n", err)
			os.Exit(1)
		}
		if stat.Mode()&os.ModeCharDevice != 0 {
			fmt.Fprintln(os.Stderr, "Error: clipboard contains an image, not text (use pbpaste > file.png to save)")
			os.Exit(1)
		}
		data, err := base64.StdEncoding.DecodeString(result.ImageData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to decode image data: %v\n", err)
			os.Exit(1)
		}
		if _, err := os.Stdout.Write(data); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	fmt.Print(result.Text)
	os.Exit(0)
}
