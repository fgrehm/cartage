package compat

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fgrehm/cartage/internal/notify"
)

// HandleNotifySend handles notify-send compatibility mode.
func HandleNotifySend(args []string) {
	for _, arg := range args[1:] {
		switch arg {
		case "-v", "--version":
			fmt.Printf("cartage (notify-send compatible)\n")
			os.Exit(0)
		case "-h", "--help":
			printNotifySendHelp()
			os.Exit(0)
		}
	}

	payload, err := parseNotifySendArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Try 'notify-send --help' for more information")
		os.Exit(1)
	}

	sendNotifyAndExit(payload)
}

// parseNotifySendArgs parses notify-send CLI arguments into a Payload.
func parseNotifySendArgs(args []string) (notify.Payload, error) {
	var title, body *string
	var icon, urgency *string
	var timeout uint32 = 5000

	i := 1 // Skip argv[0]
	for i < len(args) {
		arg := args[i]

		switch arg {
		case "-i", "--icon":
			i++
			if i < len(args) {
				icon = &args[i]
			}
		case "-t", "--expire-time":
			i++
			if i < len(args) {
				if t, err := strconv.ParseUint(args[i], 10, 32); err == nil {
					timeout = uint32(t)
				}
			}
		case "-u", "--urgency":
			i++
			if i < len(args) {
				urgency = &args[i]
			}
		case "-a", "--app-name", "-c", "--category":
			i++ // Skip next arg
		case "--hint":
			i++ // Skip next arg
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Warning: unknown option: %s\n", arg)
			} else {
				if title == nil {
					title = &args[i]
				} else if body == nil {
					body = &args[i]
				}
			}
		}
		i++
	}

	if title == nil {
		return notify.Payload{}, fmt.Errorf("title is required")
	}

	return notify.Payload{
		Title:   *title,
		Body:    body,
		Mode:    notify.ModeToast,
		Icon:    icon,
		Timeout: &timeout,
		Urgency: urgency,
	}, nil
}

func printNotifySendHelp() {
	fmt.Println("Usage: notify-send [OPTIONS] TITLE [BODY]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -i, --icon ICON          Icon name or path")
	fmt.Println("  -t, --expire-time TIME   Timeout in milliseconds")
	fmt.Println("  -u, --urgency LEVEL      Urgency (low, normal, critical)")
	fmt.Println("  -a, --app-name NAME      Application name (ignored)")
	fmt.Println("  -c, --category TYPE      Category (ignored)")
	fmt.Println("  -h, --help               Show this help")
	fmt.Println("  -v, --version            Show version")
	fmt.Println()
	fmt.Println("Note: This is cartage in notify-send compatibility mode.")
}
