package main

import (
	"os"

	"github.com/fgrehm/cartage/cli"
	"github.com/fgrehm/cartage/internal/compat"
)

func main() {
	// Check compatibility modes BEFORE Cobra parses CLI.
	// This prevents flag conflicts (e.g., --text is unknown to Cobra but valid for yad).
	mode := compat.GetCompatMode(os.Args[0])

	switch mode {
	case "notify-send":
		compat.HandleNotifySend(os.Args)
	case "yad":
		compat.HandleYad(os.Args)
	case "zenity":
		compat.HandleZenity(os.Args)
	case "kdialog":
		compat.HandleKdialog(os.Args)
	case "xdg-open":
		compat.HandleXdgOpen(os.Args)
	case "pbcopy":
		compat.HandlePbcopy(os.Args)
	case "pbpaste":
		compat.HandlePbpaste(os.Args)
	default:
		cli.Execute()
	}
}
