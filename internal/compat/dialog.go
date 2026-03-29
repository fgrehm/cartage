package compat

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fgrehm/cartage/internal/notify"
)

// HandleYad handles yad compatibility mode.
func HandleYad(args []string) {
	handleDialogCompat(args, "yad")
}

// HandleZenity handles zenity compatibility mode.
func HandleZenity(args []string) {
	handleDialogCompat(args, "zenity")
}

// handleDialogCompat is the shared implementation for yad and zenity.
func handleDialogCompat(args []string, toolName string) {
	for _, arg := range args[1:] {
		switch arg {
		case "--version":
			fmt.Printf("cartage (%s compatible)\n", toolName)
			os.Exit(0)
		case "--help":
			printDialogHelp(toolName)
			os.Exit(0)
		}
	}

	payload := parseDialogArgs(args, toolName)
	sendNotifyAndExit(payload)
}

// parseDialogArgs parses yad/zenity CLI arguments into a Payload.
// Supports both equals-separated (--text=value) and space-separated (--text value) arguments.
func parseDialogArgs(args []string, toolName string) notify.Payload {
	var title, text, icon *string
	var width, height *uint32
	var isConfirm bool
	var urgency *string

	i := 1 // Skip argv[0]
	for i < len(args) {
		arg := args[i]

		// Check for equals-separated format: --text=value
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			flag := parts[0]
			value := parts[1]

			switch flag {
			case "--title":
				title = new(value)
			case "--text":
				text = new(value)
			case "--width":
				if w, err := strconv.ParseUint(value, 10, 32); err == nil {
					w32 := uint32(w)
					width = &w32
				}
			case "--height":
				if h, err := strconv.ParseUint(value, 10, 32); err == nil {
					h32 := uint32(h)
					height = &h32
				}
			case "--image":
				icon = new(value)
			case "--timeout", "--button", "--borders", "--buttons-layout":
				// Ignore
			}
			i++
			continue
		}

		// Space-separated format: --text value
		switch arg {
		case "--info":
			urgency = new("normal")
		case "--warning":
			urgency = new("normal")
		case "--error":
			urgency = new("critical")
		case "--question":
			isConfirm = true
		case "--title":
			i++
			if i < len(args) {
				title = &args[i]
			}
		case "--text":
			i++
			if i < len(args) {
				text = &args[i]
			}
		case "--width":
			i++
			if i < len(args) {
				if w, err := strconv.ParseUint(args[i], 10, 32); err == nil {
					w32 := uint32(w)
					width = &w32
				}
			}
		case "--height":
			i++
			if i < len(args) {
				if h, err := strconv.ParseUint(args[i], 10, 32); err == nil {
					h32 := uint32(h)
					height = &h32
				}
			}
		case "--image":
			i++
			if i < len(args) {
				icon = &args[i]
			}
		case "--timeout", "--button", "--borders", "--buttons-layout":
			i++ // Skip the value
		case "--on-top", "--center", "--no-escape", "--fixed",
			"--skip-taskbar", "--selectable-labels":
			// No-op
		}
		i++
	}

	// Fallback: use text as title if no title specified
	if title == nil && text != nil {
		title = text
	}
	if title == nil {
		title = new("Dialog")
	}

	mode := notify.ModeAlert
	if isConfirm {
		mode = notify.ModeConfirm
	}

	return notify.Payload{
		Title:    *title,
		Body:     text,
		Mode:     mode,
		ToolHint: &toolName,
		Icon:     icon,
		Urgency:  urgency,
		Width:    width,
		Height:   height,
	}
}

func printDialogHelp(toolName string) {
	fmt.Printf("Usage: %s [OPTIONS]\n", toolName)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --info                   Info dialog")
	fmt.Println("  --warning                Warning dialog")
	fmt.Println("  --error                  Error dialog")
	fmt.Println("  --question               Question dialog (Yes/No)")
	fmt.Println("  --title TEXT             Dialog title")
	fmt.Println("  --text TEXT              Dialog text")
	fmt.Println("  --width N                Dialog width")
	fmt.Println("  --height N               Dialog height")
	fmt.Println("  --image NAME             Icon name")
	fmt.Println("  --help                   Show this help")
	fmt.Println("  --version                Show version")
	fmt.Println()
	fmt.Printf("Note: This is cartage in %s compatibility mode.\n", toolName)
}
