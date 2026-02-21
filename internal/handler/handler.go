package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fgrehm/cartage/internal/protocol"
)

// Handler processes a specific action type.
type Handler interface {
	// Action returns the action name this handler responds to.
	Action() string
	// Handle processes the request payload and returns a response.
	Handle(ctx context.Context, payload json.RawMessage) (*protocol.Response, error)
}

// Registry maps action names to handlers.
type Registry struct {
	handlers map[string]Handler
}

// NewRegistry creates an empty handler registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

// Register adds a handler for its declared action.
func (r *Registry) Register(h Handler) {
	r.handlers[h.Action()] = h
}

// Dispatch routes a request to the appropriate handler.
// It validates the protocol version and looks up the handler by action name.
func (r *Registry) Dispatch(ctx context.Context, req *protocol.Request) *protocol.Response {
	// Treat missing version (0) as v1
	version := req.Version
	if version == 0 {
		version = 1
	}

	if version != protocol.CurrentVersion {
		return protocol.ErrorResponse(fmt.Sprintf("unsupported protocol version: %d", req.Version))
	}

	h, ok := r.handlers[req.Action]
	if !ok {
		return protocol.ErrorResponse(fmt.Sprintf("unknown action: %s", req.Action))
	}

	resp, err := h.Handle(ctx, req.Payload)
	if err != nil {
		return protocol.ErrorResponse(err.Error())
	}

	return resp
}
