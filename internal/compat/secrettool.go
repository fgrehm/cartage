package compat

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fgrehm/cartage/internal/client"
	"github.com/fgrehm/cartage/internal/protocol"
	"github.com/fgrehm/cartage/internal/secret"
)

// HandleSecretTool handles secret-tool compatibility mode.
// It maps secret-tool subcommands onto the cartage secret action:
//
//	store  -> secret set (secret read from stdin)
//	lookup -> secret get (secret printed to stdout)
//	clear / search -> not supported (clear error)
func HandleSecretTool(args []string) {
	for _, arg := range args[1:] {
		switch arg {
		case "-h", "--help":
			printSecretToolHelp()
			os.Exit(0)
		case "-v", "--version":
			fmt.Printf("cartage (secret-tool compatible)\n")
			os.Exit(0)
		}
	}

	subcommand, attrs, err := parseSecretToolArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	service := attrs["service"]
	user := attrs["username"]

	switch subcommand {
	case "store":
		if service == "" || user == "" {
			fmt.Fprintln(os.Stderr, "Error: store requires 'service' and 'username' attributes")
			os.Exit(1)
		}
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		sendSecretAndExit(secret.Payload{
			Op:      secret.OpSet,
			Service: service,
			User:    user,
			Secret:  string(data),
		})

	case "lookup":
		if service == "" || user == "" {
			fmt.Fprintln(os.Stderr, "Error: lookup requires 'service' and 'username' attributes")
			os.Exit(1)
		}
		payloadJSON, err := json.Marshal(secret.Payload{Op: secret.OpGet, Service: service, User: user})
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
		os.Exit(0)

	case "clear", "search":
		fmt.Fprintf(os.Stderr, "Error: secret-tool %s is not supported by cartage\n", subcommand)
		os.Exit(1)

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown secret-tool subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

// sendSecretAndExit marshals a secret payload, sends it to the daemon, and exits.
func sendSecretAndExit(payload secret.Payload) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	client.MustSend(protocol.Request{
		Version: protocol.CurrentVersion,
		Action:  "secret",
		Payload: payloadJSON,
	})
	os.Exit(0)
}

// parseSecretToolArgs parses secret-tool CLI arguments into a subcommand and an
// attribute/value map. It skips options such as --label=..., --label LABEL,
// --all, and --unlock.
func parseSecretToolArgs(args []string) (string, map[string]string, error) {
	attrs := make(map[string]string)
	if len(args) < 2 {
		return "", nil, fmt.Errorf("missing subcommand (store, lookup, clear, search)")
	}
	subcommand := args[1]
	i := 2
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			if arg == "--label" {
				i += 2
				continue
			}
			i++
			continue
		}
		if i+1 >= len(args) {
			return "", nil, fmt.Errorf("missing value for attribute %q", arg)
		}
		attrs[arg] = args[i+1]
		i += 2
	}
	return subcommand, attrs, nil
}

func printSecretToolHelp() {
	fmt.Println("Usage: secret-tool COMMAND [OPTIONS] [ATTRIBUTE VALUE ...]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  store --label=LABEL ATTRIBUTE VALUE ...   Store a secret (read from stdin)")
	fmt.Println("  lookup ATTRIBUTE VALUE ...                Print a secret to stdout")
	fmt.Println("  clear ATTRIBUTE VALUE ...                 Delete a secret (not supported)")
	fmt.Println("  search ATTRIBUTE VALUE ...                Search secrets (not supported)")
	fmt.Println()
	fmt.Println("Note: This is cartage in secret-tool compatibility mode.")
}
