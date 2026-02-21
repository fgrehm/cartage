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
		return protocol.ErrorResponse("title is required"), nil
	}

	id := uuid.New().String()

	switch p.Mode {
	case ModeToast:
		if err := sendToast(p); err != nil {
			slog.Error("failed to send toast", "error", err)
			return protocol.ErrorResponse(fmt.Sprintf("failed to send notification: %v", err)), nil
		}
		slog.Info("toast notification sent", "id", id, "title", p.Title)
		return protocol.OkResponse(Result{ID: id}), nil

	case ModeAlert:
		slog.Info("showing alert dialog", "title", p.Title)
		if err := sendAlert(p); err != nil {
			slog.Error("failed to show alert", "error", err)
			return protocol.ErrorResponse(fmt.Sprintf("failed to show alert dialog: %v", err)), nil
		}
		slog.Info("alert dialog dismissed", "id", id)
		return protocol.OkResponse(Result{ID: id}), nil

	case ModeConfirm:
		slog.Info("showing confirm dialog", "title", p.Title)
		confirmed, err := sendConfirm(p)
		if err != nil {
			slog.Error("failed to show confirm dialog", "error", err)
			return protocol.ErrorResponse(fmt.Sprintf("failed to show confirm dialog: %v", err)), nil
		}
		slog.Info("confirm dialog answered", "id", id, "confirmed", confirmed)
		return protocol.OkResponse(Result{ID: id, Confirmed: &confirmed}), nil

	default:
		return protocol.ErrorResponse(fmt.Sprintf("unknown notification mode: %s", p.Mode)), nil
	}
}
