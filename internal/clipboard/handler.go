package clipboard

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/fgrehm/cartage/internal/protocol"
)

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
		return Result{}, fmt.Errorf("failed to decode clipboard result: %w", err)
	}
	return r, nil
}

// Op represents the clipboard operation type.
type Op string

const (
	OpRead  Op = "read"
	OpWrite Op = "write"
)

// Payload is the action-specific data for a "clipboard" request.
type Payload struct {
	Op        Op      `json:"op"`
	Text      *string `json:"text,omitempty"`       // for write: text content
	ImageData *string `json:"image_data,omitempty"` // for write: base64-encoded image
	ImageType *string `json:"image_type,omitempty"` // "png", "jpeg" (default: "png")
}

// ContentType represents what type of content is in the clipboard.
type ContentType string

const (
	ContentText  ContentType = "text"
	ContentImage ContentType = "image"
)

// Result is the action-specific data returned in Response.Data for clipboard read.
type Result struct {
	ContentType ContentType `json:"content_type"`
	Text        string      `json:"text,omitempty"`
	ImageData   string      `json:"image_data,omitempty"` // base64
	ImageType   string      `json:"image_type,omitempty"` // "png", "jpeg"
}

// Handler implements handler.Handler for the "clipboard" action.
type Handler struct{}

func (h *Handler) Action() string { return "clipboard" }

func (h *Handler) Handle(ctx context.Context, raw json.RawMessage) (*protocol.Response, error) {
	var p Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return protocol.ErrorResponse(fmt.Sprintf("invalid clipboard payload: %v", err)), nil
	}

	switch p.Op {
	case OpRead:
		tool, err := detectClipboardTool()
		if err != nil {
			return protocol.ErrorResponse(err.Error()), nil
		}
		return h.handleRead(ctx, tool)

	case OpWrite:
		if p.Text == nil && p.ImageData == nil {
			return protocol.ErrorResponse("write requires exactly one of: text, image_data"), nil
		}
		if p.Text != nil && p.ImageData != nil {
			return protocol.ErrorResponse("write requires exactly one of: text, image_data"), nil
		}
		tool, err := detectClipboardTool()
		if err != nil {
			return protocol.ErrorResponse(err.Error()), nil
		}
		return h.handleWrite(ctx, tool, p)

	default:
		return protocol.ErrorResponse(fmt.Sprintf("unknown clipboard op: %s", p.Op)), nil
	}
}

func (h *Handler) handleRead(ctx context.Context, tool ClipboardTool) (*protocol.Response, error) {
	types, err := listClipboardTypes(tool)
	if err != nil {
		return protocol.ErrorResponse(fmt.Sprintf("failed to list clipboard types: %v", err)), nil
	}

	// Prefer text over image
	for _, t := range types {
		switch t {
		case "text/plain", "text/plain;charset=utf-8", "TEXT", "STRING", "UTF8_STRING":
			text, err := readClipboardText(ctx, tool)
			if err != nil {
				return protocol.ErrorResponse(fmt.Sprintf("failed to read clipboard text: %v", err)), nil
			}
			slog.Info("clipboard read", "content_type", "text")
			return protocol.OkResponse(Result{ContentType: ContentText, Text: text}), nil
		}
	}

	// Try image types in preference order
	for _, mimeType := range []string{"image/png", "image/jpeg"} {
		for _, t := range types {
			if t == mimeType {
				data, err := readClipboardImage(ctx, tool, mimeType)
				if err != nil {
					return protocol.ErrorResponse(fmt.Sprintf("failed to read clipboard image: %v", err)), nil
				}
				imgType := "png"
				if mimeType == "image/jpeg" {
					imgType = "jpeg"
				}
				encoded := base64.StdEncoding.EncodeToString(data)
				slog.Info("clipboard read", "content_type", "image", "image_type", imgType)
				return protocol.OkResponse(Result{
					ContentType: ContentImage,
					ImageData:   encoded,
					ImageType:   imgType,
				}), nil
			}
		}
	}

	// Fallback: attempt to read as text
	text, err := readClipboardText(ctx, tool)
	if err != nil {
		return protocol.ErrorResponse("clipboard is empty or contains unsupported content"), nil
	}
	slog.Info("clipboard read", "content_type", "text", "fallback", true)
	return protocol.OkResponse(Result{ContentType: ContentText, Text: text}), nil
}

func (h *Handler) handleWrite(ctx context.Context, tool ClipboardTool, p Payload) (*protocol.Response, error) {
	if p.Text != nil {
		if err := writeClipboardText(ctx, tool, *p.Text); err != nil {
			return protocol.ErrorResponse(fmt.Sprintf("failed to write text to clipboard: %v", err)), nil
		}
		slog.Info("clipboard write", "content_type", "text")
		return protocol.OkResponse(nil), nil
	}

	// Image write
	data, err := base64.StdEncoding.DecodeString(*p.ImageData)
	if err != nil {
		return protocol.ErrorResponse(fmt.Sprintf("invalid base64 image_data: %v", err)), nil
	}

	imgType := "png"
	if p.ImageType != nil && *p.ImageType != "" {
		imgType = *p.ImageType
	}
	mimeType := "image/" + imgType

	if err := writeClipboardImage(ctx, tool, data, mimeType); err != nil {
		return protocol.ErrorResponse(fmt.Sprintf("failed to write image to clipboard: %v", err)), nil
	}
	slog.Info("clipboard write", "content_type", "image", "image_type", imgType)
	return protocol.OkResponse(nil), nil
}
