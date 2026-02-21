package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cartage",
	Short: "Container-to-host bridge daemon",
	Long: `cartage routes intents (notifications, xdg-open, and more) from containers
to the host desktop. One socket, one daemon, one binary.`,
	Version: version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(clipboardCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(notifyCmd)
	rootCmd.AddCommand(openCmd)
}
