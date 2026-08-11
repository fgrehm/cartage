package secret

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fgrehm/cartage/internal/protocol"
	"github.com/zalando/go-keyring"
)

// Op represents the secret operation type.
type Op string

const (
	OpGet Op = "get"
)

// Payload is the action-specific data for a "secret" request.
type Payload struct {
	Op      Op     `json:"op"`
	Service string `json:"service"`
	User    string `json:"user"`
}

// Result is the action-specific data returned in Response.Data for a get.
type Result struct {
	Secret string `json:"secret"`
}

// ParseResult decodes daemon response data into a typed Result.
// Response.Data is an any that, after JSON round-trip, becomes map[string]any;
// this re-marshals and unmarshals it to get a properly typed struct.
func ParseResult(data any) (Result, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return Result{}, fmt.Errorf("failed to encode response data: %w", err)
	}
	var r Result
	if err := json.Unmarshal(b, &r); err != nil {
		return Result{}, fmt.Errorf("failed to decode secret result: %w", err)
	}
	return r, nil
}

// Handler implements handler.Handler for the "secret" action.
// It retrieves secrets from the host OS keychain.
type Handler struct{}

func (h *Handler) Action() string { return "secret" }

func (h *Handler) Handle(ctx context.Context, raw json.RawMessage) (*protocol.Response, error) {
	var p Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return protocol.ErrorResponse(fmt.Sprintf("invalid secret payload: %v", err)), nil
	}

	switch p.Op {
	case OpGet:
		return h.handleGet(p)

	default:
		return protocol.ErrorResponse(fmt.Sprintf("secret: unknown op: %s", p.Op)), nil
	}
}

func (h *Handler) handleGet(p Payload) (*protocol.Response, error) {
	if p.Service == "" {
		return protocol.ErrorResponse("secret: get requires service"), nil
	}
	if p.User == "" {
		return protocol.ErrorResponse("secret: get requires user"), nil
	}

	secret, err := keyring.Get(p.Service, p.User)
	if err != nil {
		return protocol.ErrorResponse(fmt.Sprintf("secret: failed to get %q for %q: %v", p.Service, p.User, err)), nil
	}

	return protocol.OkResponse(Result{Secret: secret}), nil
}
