package compat

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fgrehm/cartage/internal/notify"
)

// HandleKdialog handles kdialog compatibility mode.
func HandleKdialog(args []string) {
	var title, text, icon *string
	var isConfirm bool
	var isToast bool
	var urgency *string
	var timeout *uint32

	i := 1 // Skip argv[0]
	for i < len(args) {
		arg := args[i]

		switch arg {
		case "--version":
			fmt.Printf("cartage (kdialog compatible)\n")
			os.Exit(0)
		case "--help":
			printKdialogHelp()
			os.Exit(0)
		case "--passivepopup":
			isToast = true
			i++
			if i < len(args) {
				text = &args[i]
			}
			// Next positional arg is timeout in seconds
			i++
			if i < len(args) {
				if sec, err := strconv.ParseUint(args[i], 10, 32); err == nil {
					ms := uint32(sec * 1000)
					timeout = &ms
				}
			}
		case "--msgbox":
			urgency = strPtr("normal")
			i++
			if i < len(args) {
				text = &args[i]
			}
		case "--error":
			urgency = strPtr("critical")
			i++
			if i < len(args) {
				text = &args[i]
			}
		case "--yesno":
			isConfirm = true
			i++
			if i < len(args) {
				text = &args[i]
			}
		case "--title":
			i++
			if i < len(args) {
				title = &args[i]
			}
		case "--icon":
			i++
			if i < len(args) {
				icon = &args[i]
			}
		}
		i++
	}

	// Fallback: use text as title if no title
	if title == nil && text != nil {
		title = text
	}
	if title == nil {
		title = strPtr("Dialog")
	}

	mode := notify.ModeAlert
	if isToast {
		mode = notify.ModeToast
	} else if isConfirm {
		mode = notify.ModeConfirm
	}

	toolHint := "kdialog"
	payload := notify.Payload{
		Title:    *title,
		Body:     text,
		Mode:     mode,
		ToolHint: &toolHint,
		Icon:     icon,
		Urgency:  urgency,
		Timeout:  timeout,
	}

	sendNotifyAndExit(payload)
}

func printKdialogHelp() {
	fmt.Println("Usage: kdialog [OPTIONS]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --passivepopup TEXT SEC  Passive popup notification")
	fmt.Println("  --msgbox TEXT            Message box")
	fmt.Println("  --error TEXT             Error dialog")
	fmt.Println("  --yesno TEXT             Yes/No dialog")
	fmt.Println("  --title TEXT             Dialog title")
	fmt.Println("  --icon NAME              Icon name or path")
	fmt.Println("  --help                   Show this help")
	fmt.Println("  --version                Show version")
	fmt.Println()
	fmt.Println("Note: This is cartage in kdialog compatibility mode.")
}
