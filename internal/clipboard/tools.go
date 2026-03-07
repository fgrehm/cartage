package clipboard

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// ClipboardTool represents which clipboard tool is available.
type ClipboardTool int

const (
	ToolWlClipboard ClipboardTool = iota
	ToolXclip
	ToolXsel
	ToolNone
)

var clipboardToolOnce sync.Once
var clipboardToolCached ClipboardTool
var clipboardToolErr error

// detectClipboardTool returns the best available clipboard tool.
// Results are cached after the first call.
//
// Detection order:
//  1. $WAYLAND_DISPLAY set + wl-copy in PATH → wl-clipboard
//  2. xclip in PATH → xclip
//  3. xsel in PATH → xsel
//  4. wl-copy in PATH (no env check) → wl-clipboard
//  5. error: no tool available
func detectClipboardTool() (ClipboardTool, error) {
	clipboardToolOnce.Do(func() {
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			if _, err := exec.LookPath("wl-copy"); err == nil {
				clipboardToolCached = ToolWlClipboard
				return
			}
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			clipboardToolCached = ToolXclip
			return
		}
		if _, err := exec.LookPath("xsel"); err == nil {
			clipboardToolCached = ToolXsel
			return
		}
		if _, err := exec.LookPath("wl-copy"); err == nil {
			clipboardToolCached = ToolWlClipboard
			return
		}
		clipboardToolCached = ToolNone
		clipboardToolErr = fmt.Errorf("no clipboard tool available (install wl-clipboard, xclip, or xsel)")
	})
	return clipboardToolCached, clipboardToolErr
}

// listClipboardTypes returns the MIME types currently available in the clipboard.
func listClipboardTypes(ctx context.Context, tool ClipboardTool) ([]string, error) {
	var cmd *exec.Cmd
	switch tool {
	case ToolWlClipboard:
		cmd = exec.CommandContext(ctx, "wl-paste", "--list-types")
	case ToolXclip:
		cmd = exec.CommandContext(ctx, "xclip", "-t", "TARGETS", "-selection", "clipboard", "-o")
	case ToolXsel:
		// xsel does not support listing types; assume text
		return []string{"text/plain"}, nil
	default:
		return nil, fmt.Errorf("no clipboard tool available")
	}

	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("listing clipboard types: %w", ctx.Err())
		}
		// An empty clipboard is not a fatal error
		return nil, nil
	}

	var types []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			types = append(types, t)
		}
	}
	return types, nil
}

// readClipboardText reads the clipboard content as plain text.
func readClipboardText(ctx context.Context, tool ClipboardTool) (string, error) {
	var cmd *exec.Cmd
	switch tool {
	case ToolWlClipboard:
		cmd = exec.CommandContext(ctx, "wl-paste", "--no-newline", "--type", "text/plain")
	case ToolXclip:
		cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-o")
	case ToolXsel:
		cmd = exec.CommandContext(ctx, "xsel", "--clipboard", "--output")
	default:
		return "", fmt.Errorf("no clipboard tool available")
	}

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("clipboard read failed: %w", err)
	}
	return string(out), nil
}

// readClipboardImage reads the clipboard content as raw image bytes.
func readClipboardImage(ctx context.Context, tool ClipboardTool, mimeType string) ([]byte, error) {
	var cmd *exec.Cmd
	switch tool {
	case ToolWlClipboard:
		cmd = exec.CommandContext(ctx, "wl-paste", "--type", mimeType)
	case ToolXclip:
		cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-t", mimeType, "-o")
	case ToolXsel:
		return nil, fmt.Errorf("xsel does not support image clipboard operations")
	default:
		return nil, fmt.Errorf("no clipboard tool available")
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("clipboard image read failed: %w", err)
	}
	return out, nil
}

// writeClipboardText writes text to the clipboard via stdin.
func writeClipboardText(ctx context.Context, tool ClipboardTool, text string) error {
	var cmd *exec.Cmd
	switch tool {
	case ToolWlClipboard:
		cmd = exec.CommandContext(ctx, "wl-copy")
	case ToolXclip:
		cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard")
	case ToolXsel:
		cmd = exec.CommandContext(ctx, "xsel", "--clipboard", "--input")
	default:
		return fmt.Errorf("no clipboard tool available")
	}

	cmd.Stdin = bytes.NewBufferString(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard write failed: %w", err)
	}
	return nil
}

// writeClipboardImage writes image data to the clipboard via stdin.
// xsel does not support image operations and returns an error.
func writeClipboardImage(ctx context.Context, tool ClipboardTool, data []byte, mimeType string) error {
	var cmd *exec.Cmd
	switch tool {
	case ToolWlClipboard:
		cmd = exec.CommandContext(ctx, "wl-copy", "--type", mimeType)
	case ToolXclip:
		cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-t", mimeType)
	case ToolXsel:
		return fmt.Errorf("xsel does not support image clipboard operations")
	default:
		return fmt.Errorf("no clipboard tool available")
	}

	cmd.Stdin = bytes.NewReader(data)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard image write failed: %w", err)
	}
	return nil
}
