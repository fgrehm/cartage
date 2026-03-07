package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/fgrehm/cartage/internal/protocol"
	"github.com/google/uuid"
)

// Handler implements handler.Handler for the "notify" action.
type Handler struct{}

func (h *Handler) Action() string { return "notify" }

func (h *Handler) Handle(ctx context.Context, raw json.RawMessage) (*protocol.Response, error) {
	var p Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return protocol.ErrorResponse(fmt.Sprintf("invalid notify payload: %v", err)), nil
	}

	if p.Title == "" {
		return protocol.ErrorResponse("notify: title is required"), nil
	}

	id := uuid.New().String()

	switch p.Mode {
	case ModeToast:
		if err := sendToast(ctx, p); err != nil {
			slog.Error("failed to send toast", "error", err)
			return protocol.ErrorResponse(fmt.Sprintf("notify: failed to send notification: %v", err)), nil
		}
		slog.Info("toast notification sent", "id", id, "title", p.Title)
		return protocol.OkResponse(Result{ID: id}), nil

	case ModeAlert:
		slog.Info("showing alert dialog", "title", p.Title)
		if err := sendAlert(ctx, p); err != nil {
			slog.Error("failed to show alert", "error", err)
			return protocol.ErrorResponse(fmt.Sprintf("notify: failed to show alert dialog: %v", err)), nil
		}
		slog.Info("alert dialog dismissed", "id", id)
		return protocol.OkResponse(Result{ID: id}), nil

	case ModeConfirm:
		slog.Info("showing confirm dialog", "title", p.Title)
		confirmed, err := sendConfirm(ctx, p)
		if err != nil {
			slog.Error("failed to show confirm dialog", "error", err)
			return protocol.ErrorResponse(fmt.Sprintf("notify: failed to show confirm dialog: %v", err)), nil
		}
		slog.Info("confirm dialog answered", "id", id, "confirmed", confirmed)
		return protocol.OkResponse(Result{ID: id, Confirmed: &confirmed}), nil

	default:
		return protocol.ErrorResponse(fmt.Sprintf("notify: unknown mode: %s", p.Mode)), nil
	}
}
