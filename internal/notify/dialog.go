package notify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
)

// DialogTool represents available dialog tools on the system.
type DialogTool int

const (
	DialogToolYad DialogTool = iota
	DialogToolZenity
	DialogToolKdialog
	DialogToolNone
)

// ToastTool represents available toast notification tools on the system.
type ToastTool int

const (
	ToastToolNotifySend ToastTool = iota
	ToastToolKdialog
	ToastToolNone
)

// toolAvailableCache caches exec.LookPath results to avoid repeated PATH lookups.
var toolAvailableCache sync.Map

func isToolAvailable(name string) bool {
	if v, ok := toolAvailableCache.Load(name); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	_, err := exec.LookPath(name)
	available := err == nil
	toolAvailableCache.Store(name, available)
	return available
}

// detectToastTool finds which toast notification tool is available.
// If hint matches an available tool, that tool is preferred.
// Otherwise falls back to the default order: notify-send > kdialog.
func detectToastTool(hint *string) ToastTool {
	if hint != nil && *hint == "kdialog" {
		if isToolAvailable("kdialog") {
			return ToastToolKdialog
		}
	}
	if isToolAvailable("notify-send") {
		return ToastToolNotifySend
	}
	if isToolAvailable("kdialog") {
		return ToastToolKdialog
	}
	return ToastToolNone
}

// detectDialogTool finds which dialog tool is available on the system.
// If hint matches an available tool, that tool is preferred.
// Otherwise falls back to the default order: yad > zenity > kdialog.
func detectDialogTool(hint *string) DialogTool {
	if hint != nil {
		switch *hint {
		case "kdialog":
			if isToolAvailable("kdialog") {
				return DialogToolKdialog
			}
		case "zenity":
			if isToolAvailable("zenity") {
				return DialogToolZenity
			}
		case "yad":
			if isToolAvailable("yad") {
				return DialogToolYad
			}
		}
	}
	if isToolAvailable("yad") {
		return DialogToolYad
	}
	if isToolAvailable("zenity") {
		return DialogToolZenity
	}
	if isToolAvailable("kdialog") {
		return DialogToolKdialog
	}
	return DialogToolNone
}

// dialogParams holds the resolved display values for a dialog.
type dialogParams struct {
	title  string
	text   string
	width  uint32
	height uint32
}

// newDialogParams resolves the display title, text, width, and height from a
// Payload, using fallbackTitle when the payload has no body.
func newDialogParams(p Payload, fallbackTitle string) dialogParams {
	title := p.Title
	text := ""
	if p.Body != nil {
		text = *p.Body
	} else {
		title = fallbackTitle
		text = p.Title
	}
	width := uint32(400)
	if p.Width != nil {
		width = *p.Width
	}
	height := uint32(150)
	if p.Height != nil {
		height = *p.Height
	}
	return dialogParams{title: title, text: text, width: width, height: height}
}

// sendAlert shows a blocking alert dialog with an OK button.
func sendAlert(ctx context.Context, p Payload) error {
	tool := detectDialogTool(p.ToolHint)
	if tool == DialogToolNone {
		return fmt.Errorf("no dialog tool available (install yad, zenity, or kdialog)")
	}

	dp := newDialogParams(p, "Alert")

	var cmd *exec.Cmd

	switch tool {
	case DialogToolYad:
		cmd = exec.CommandContext(ctx, "yad")
		switch {
		case p.Urgency != nil && *p.Urgency == "critical":
			cmd.Args = append(cmd.Args, "--error")
		default:
			cmd.Args = append(cmd.Args, "--info")
		}
		cmd.Args = append(cmd.Args,
			"--title", dp.title,
			"--text", dp.text,
			fmt.Sprintf("--width=%d", dp.width),
			fmt.Sprintf("--height=%d", dp.height),
			"--center",
			"--button=OK:0",
			"--on-top",
			"--fixed",
			"--borders=15",
			"--skip-taskbar",
			"--buttons-layout=center",
		)
		if p.Icon != nil {
			cmd.Args = append(cmd.Args, "--image", *p.Icon)
		}

	case DialogToolZenity:
		cmd = exec.CommandContext(ctx, "zenity")
		switch {
		case p.Urgency != nil && *p.Urgency == "critical":
			cmd.Args = append(cmd.Args, "--error")
		default:
			cmd.Args = append(cmd.Args, "--info")
		}
		cmd.Args = append(cmd.Args,
			"--title", dp.title,
			"--text", dp.text,
			fmt.Sprintf("--width=%d", dp.width),
		)

	case DialogToolKdialog:
		cmd = exec.CommandContext(ctx, "kdialog")
		switch {
		case p.Urgency != nil && *p.Urgency == "critical":
			cmd.Args = append(cmd.Args, "--error", dp.text)
		default:
			cmd.Args = append(cmd.Args, "--msgbox", dp.text)
		}
		cmd.Args = append(cmd.Args, "--title", dp.title)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if trimmed := strings.TrimSpace(string(output)); len(trimmed) > 0 {
				return fmt.Errorf("dialog tool failed (exit %d): %s", exitErr.ExitCode(), trimmed)
			}
			slog.Debug("alert dialog dismissed", "exit_code", exitErr.ExitCode())
			return nil
		}
		return fmt.Errorf("failed to execute dialog tool: %w", err)
	}

	slog.Debug("alert dialog dismissed successfully")
	return nil
}

// sendConfirm shows a blocking confirm dialog with Yes/No buttons.
// Returns true for Yes, false for No/Cancel.
func sendConfirm(ctx context.Context, p Payload) (bool, error) {
	tool := detectDialogTool(p.ToolHint)
	if tool == DialogToolNone {
		return false, fmt.Errorf("no dialog tool available (install yad, zenity, or kdialog)")
	}

	dp := newDialogParams(p, "Confirmation")

	var cmd *exec.Cmd

	switch tool {
	case DialogToolYad:
		cmd = exec.CommandContext(ctx, "yad",
			"--question",
			"--title", dp.title,
			"--text", dp.text,
			fmt.Sprintf("--width=%d", dp.width),
			fmt.Sprintf("--height=%d", dp.height),
			"--center",
			"--on-top",
			"--fixed",
			"--borders=15",
			"--skip-taskbar",
			"--buttons-layout=center",
			"--no-escape",
		)
		if p.Icon != nil {
			cmd.Args = append(cmd.Args, "--image", *p.Icon)
		}

	case DialogToolZenity:
		cmd = exec.CommandContext(ctx, "zenity",
			"--question",
			"--title", dp.title,
			"--text", dp.text,
			fmt.Sprintf("--width=%d", dp.width),
		)

	case DialogToolKdialog:
		cmd = exec.CommandContext(ctx, "kdialog",
			"--yesno", dp.text,
			"--title", dp.title,
		)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if trimmed := strings.TrimSpace(string(output)); len(trimmed) > 0 {
				return false, fmt.Errorf("dialog tool failed (exit %d): %s", exitErr.ExitCode(), trimmed)
			}
			exitCode := exitErr.ExitCode()
			switch exitCode {
			case 1:
				return false, nil
			default:
				slog.Warn("dialog returned unexpected exit code", "code", exitCode)
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to execute dialog tool: %w", err)
	}

	return true, nil
}
