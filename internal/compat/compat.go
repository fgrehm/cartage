package compat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fgrehm/cartage/internal/client"
	"github.com/fgrehm/cartage/internal/notify"
	"github.com/fgrehm/cartage/internal/protocol"
)

// GetCompatMode returns which compatibility mode to use, or empty string for normal mode.
func GetCompatMode(programName string) string {
	base := filepath.Base(programName)
	switch {
	case strings.HasPrefix(base, "notify-send"):
		return "notify-send"
	case strings.HasPrefix(base, "yad"):
		return "yad"
	case strings.HasPrefix(base, "zenity"):
		return "zenity"
	case strings.HasPrefix(base, "kdialog"):
		return "kdialog"
	case strings.HasPrefix(base, "xdg-open"):
		return "xdg-open"
	case strings.HasPrefix(base, "pbcopy"):
		return "pbcopy"
	case strings.HasPrefix(base, "pbpaste"):
		return "pbpaste"
	default:
		return ""
	}
}

// sendNotifyAndExit marshals a notify payload, sends it to the daemon,
// and exits with the appropriate code (for confirm dialogs, exit 1 = "No").
func sendNotifyAndExit(payload notify.Payload) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	resp := client.MustSend(protocol.Request{
		Version: protocol.CurrentVersion,
		Action:  "notify",
		Payload: payloadJSON,
	})

	if payload.Mode == notify.ModeConfirm {
		if notify.ExtractConfirmed(resp.Data) {
			os.Exit(0)
		}
		os.Exit(1)
	}

	os.Exit(0)
}
