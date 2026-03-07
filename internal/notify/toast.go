package notify

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// sendToast sends a non-blocking toast notification.
// It prefers notify-send, falling back to kdialog --passivepopup on KDE systems.
func sendToast(ctx context.Context, p Payload) error {
	tool := detectToastTool(p.ToolHint)
	switch tool {
	case ToastToolNotifySend:
		return sendToastNotifySend(ctx, p)
	case ToastToolKdialog:
		return sendToastKdialog(ctx, p)
	default:
		return fmt.Errorf("no toast notification tool available (install libnotify-bin or kdialog)")
	}
}

// sendToastNotifySend sends a toast via notify-send.
func sendToastNotifySend(ctx context.Context, p Payload) error {
	cmd := exec.CommandContext(ctx, "notify-send")
	cmd.Args = append(cmd.Args, p.Title)

	if p.Body != nil {
		cmd.Args = append(cmd.Args, *p.Body)
	}

	// icon_data takes precedence over icon
	if p.IconData != nil {
		path, err := decodeAndSaveIcon(*p.IconData, p.IconType)
		if err != nil {
			slog.Warn("failed to decode icon data", "error", err)
		} else {
			cmd.Args = append(cmd.Args, "-i", path)
			defer func() { _ = os.Remove(path) }()
		}
	} else if p.Icon != nil {
		cmd.Args = append(cmd.Args, "-i", *p.Icon)
	}

	if p.Urgency != nil {
		cmd.Args = append(cmd.Args, "-u", *p.Urgency)
	}

	if p.Timeout != nil {
		cmd.Args = append(cmd.Args, "-t", fmt.Sprint(*p.Timeout))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("notify-send failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// sendToastKdialog sends a toast via kdialog --passivepopup.
func sendToastKdialog(ctx context.Context, p Payload) error {
	text := p.Title
	if p.Body != nil {
		text = fmt.Sprintf("%s\n%s", p.Title, *p.Body)
	}

	// kdialog --passivepopup timeout is in seconds, payload is in milliseconds
	timeoutSec := 5
	if p.Timeout != nil {
		timeoutSec = int(*p.Timeout / 1000)
		if timeoutSec < 1 {
			timeoutSec = 1
		}
	}

	cmd := exec.CommandContext(ctx, "kdialog",
		"--passivepopup", text,
		fmt.Sprint(timeoutSec),
	)

	if p.Title != "" {
		cmd.Args = append(cmd.Args, "--title", p.Title)
	}

	if p.IconData != nil {
		path, err := decodeAndSaveIcon(*p.IconData, p.IconType)
		if err != nil {
			slog.Warn("failed to decode icon data", "error", err)
		} else {
			cmd.Args = append(cmd.Args, "--icon", path)
			defer func() { _ = os.Remove(path) }()
		}
	} else if p.Icon != nil {
		cmd.Args = append(cmd.Args, "--icon", *p.Icon)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kdialog --passivepopup failed: %w\nOutput: %s", err, output)
	}

	return nil
}
