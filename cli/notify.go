package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fgrehm/cartage/internal/client"
	"github.com/fgrehm/cartage/internal/notify"
	"github.com/fgrehm/cartage/internal/protocol"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "notify TITLE [BODY]",
	Short: "Send a notification to the daemon",
	Long: `Send a notification to the running daemon.

Examples:
  cartage notify "Build Complete"
  cartage notify "Build Complete" "Your project compiled successfully"
  cartage notify --alert "Error" "Build failed"
  cartage notify --confirm "Continue?" "Do you want to proceed?"
  cartage notify --icon dialog-warning "Warning" "Low disk space"`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		var body *string
		if len(args) > 1 {
			body = &args[1]
		}

		alert, _ := cmd.Flags().GetBool("alert")
		confirm, _ := cmd.Flags().GetBool("confirm")
		icon, _ := cmd.Flags().GetString("icon")
		timeout, _ := cmd.Flags().GetUint32("timeout")
		urgency, _ := cmd.Flags().GetString("urgency")

		mode := notify.ModeToast
		if alert {
			mode = notify.ModeAlert
		} else if confirm {
			mode = notify.ModeConfirm
		}

		payload := notify.Payload{
			Title:   title,
			Body:    body,
			Mode:    mode,
			Icon:    stringPtr(icon),
			Timeout: &timeout,
			Urgency: stringPtr(urgency),
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		resp, err := client.Send(protocol.Request{
			Version: protocol.CurrentVersion,
			Action:  "notify",
			Payload: payloadJSON,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Handle confirm dialog exit codes
		if confirm {
			confirmed := notify.ExtractConfirmed(resp.Data)
			if confirmed {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}

		os.Exit(0)
	},
}

func init() {
	notifyCmd.Flags().Bool("alert", false, "Show as blocking alert dialog (OK button)")
	notifyCmd.Flags().Bool("confirm", false, "Show as blocking confirm dialog (Yes/No)")
	notifyCmd.Flags().StringP("icon", "i", "", "Icon name or path")
	notifyCmd.Flags().Uint32P("timeout", "t", 5000, "Timeout in milliseconds (toast only)")
	notifyCmd.Flags().StringP("urgency", "u", "", "Urgency level (low, normal, critical)")

	notifyCmd.MarkFlagsMutuallyExclusive("alert", "confirm")
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
