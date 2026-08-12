package cli

import (
	"encoding/json"
	"fmt"
	"io"
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

var secretListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets in the host OS keychain",
	Long: `List the service/user pairs of secrets stored in the host OS keychain.

Only supported on platforms with the Secret Service dbus interface (Linux).

Examples:
  cartage secret list`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		payload := secret.Payload{Op: secret.OpList}
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

		result, err := secret.ParseListResult(resp.Data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		for _, ref := range result.Secrets {
			fmt.Printf("%s\t%s\n", ref.Service, ref.User)
		}
	},
}

var secretSetCmd = &cobra.Command{
	Use:   "set SERVICE USER [SECRET]",
	Short: "Store a secret in the host OS keychain",
	Long: `Store a secret in the host OS keychain under a service and user (account)
pair. If SECRET is omitted, it is read from stdin.

Examples:
  cartage secret set myapp alice "hunter2"
  echo -n "hunter2" | cartage secret set myapp alice`,
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		var secretValue string
		if len(args) == 3 {
			secretValue = args[2]
		} else {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
			secretValue = string(data)
		}

		payload := secret.Payload{Op: secret.OpSet, Service: args[0], User: args[1], Secret: secretValue}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		_, err = client.Send(protocol.Request{
			Version: protocol.CurrentVersion,
			Action:  "secret",
			Payload: payloadJSON,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	secretCmd.AddCommand(secretGetCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretSetCmd)
}
