package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fgrehm/cartage/internal/client"
	"github.com/fgrehm/cartage/internal/open"
	"github.com/fgrehm/cartage/internal/protocol"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open URI",
	Short: "Open a URI on the host via xdg-open",
	Long: `Forward a URI to the host's xdg-open.

Examples:
  cartage open https://example.com
  cartage open /path/to/file.pdf`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := open.MapPath(args[0])
		payload := open.Payload{URI: uri}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		_, err = client.Send(protocol.Request{
			Version: protocol.CurrentVersion,
			Action:  "open",
			Payload: payloadJSON,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}
