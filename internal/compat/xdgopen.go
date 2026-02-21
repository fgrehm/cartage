package compat

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fgrehm/cartage/internal/client"
	"github.com/fgrehm/cartage/internal/open"
	"github.com/fgrehm/cartage/internal/protocol"
)

// HandleXdgOpen handles xdg-open compatibility mode.
func HandleXdgOpen(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: xdg-open URI")
		os.Exit(1)
	}

	uri := open.MapPath(open.ResolvePath(args[1]))

	payload := open.Payload{URI: uri}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Send(protocol.Request{
		Version: protocol.CurrentVersion,
		Action:  "open",
		Payload: payloadJSON,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Status == "error" {
		fmt.Fprintf(os.Stderr, "xdg-open failed: %s\n", resp.Error)
		os.Exit(1)
	}

	os.Exit(0)
}
