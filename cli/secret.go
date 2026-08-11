package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fgrehm/cartage/internal/client"
	"github.com/fgrehm/cartage/internal/protocol"
	"github.com/fgrehm/cartage/internal/secret"
	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Access secrets stored in the host OS keychain",
}

var secretGetCmd = &cobra.Command{
	Use:   "get SERVICE USER",
	Short: "Retrieve a secret from the host OS keychain",
	Long: `Retrieve a secret stored in the host OS keychain.

The secret is looked up by its service and user (account) names, matching how
the keychain addresses entries.

Examples:
  cartage secret get myapp alice`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		payload := secret.Payload{Op: secret.OpGet, Service: args[0], User: args[1]}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		resp, err := client.Send(protocol.Request{
			Version: protocol.CurrentVersion,
			Action:  "secret",
			Payload: payloadJSON,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		result, err := secret.ParseResult(resp.Data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(result.Secret)
	},
}

func init() {
	secretCmd.AddCommand(secretGetCmd)
}
