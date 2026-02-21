package cli

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/fgrehm/cartage/internal/clipboard"
	"github.com/fgrehm/cartage/internal/handler"
	"github.com/fgrehm/cartage/internal/notify"
	"github.com/fgrehm/cartage/internal/open"
	"github.com/fgrehm/cartage/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the bridge daemon",
	Long: `Start the bridge daemon that listens for requests on a Unix socket.

The daemon will create a socket at $XDG_RUNTIME_DIR/cartage.sock by default,
or use the path specified by --socket or CARTAGE_SOCKET.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		socketPath, _ := cmd.Flags().GetString("socket")
		verbose, _ := cmd.Flags().GetBool("verbose")

		registry := handler.NewRegistry()
		registry.Register(&clipboard.Handler{})
		registry.Register(&notify.Handler{})
		registry.Register(&open.Handler{})

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		return server.Run(ctx, socketPath, registry, verbose)
	},
}

func init() {
	serveCmd.Flags().StringP("socket", "s", "", "Override socket path (default: $XDG_RUNTIME_DIR/cartage.sock)")
	serveCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
}
