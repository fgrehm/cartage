package compat

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fgrehm/cartage/internal/client"
	"github.com/fgrehm/cartage/internal/clipboard"
	"github.com/fgrehm/cartage/internal/protocol"
)

// HandlePbcopy handles pbcopy compatibility mode.
// It reads all stdin as text and sends a clipboard write request.
func HandlePbcopy(_ []string) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	text := string(data)
	payload := clipboard.Payload{
		Op:   clipboard.OpWrite,
		Text: &text,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client.MustSend(protocol.Request{
		Version: protocol.CurrentVersion,
		Action:  "clipboard",
		Payload: payloadJSON,
	})
	os.Exit(0)
}
