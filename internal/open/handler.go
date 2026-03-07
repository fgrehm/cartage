package open

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/fgrehm/cartage/internal/protocol"
)

// Payload is the action-specific data for an "open" request.
type Payload struct {
	URI string `json:"uri"`
}

// Handler implements handler.Handler for the "open" action.
// It calls xdg-open on the host to open URIs.
type Handler struct{}

func (h *Handler) Action() string { return "open" }

func (h *Handler) Handle(ctx context.Context, raw json.RawMessage) (*protocol.Response, error) {
	var p Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return protocol.ErrorResponse(fmt.Sprintf("invalid open payload: %v", err)), nil
	}

	if p.URI == "" {
		return protocol.ErrorResponse("open: uri is required"), nil
	}

	cmd := exec.CommandContext(ctx, "xdg-open", p.URI)
	if err := cmd.Run(); err != nil {
		return protocol.ErrorResponse(fmt.Sprintf("open: xdg-open failed: %v", err)), nil
	}

	return protocol.OkResponse(nil), nil
}
