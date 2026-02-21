package client

import (
	"os"
	"path/filepath"
)

// containerSocketPath is the well-known container mount point for the socket.
// Declared as a var to allow testing.
var containerSocketPath = "/run/host/cartage.sock"

// FindSocketPath discovers the Unix socket path for the daemon.
// It checks locations in priority order and returns the first that exists,
// or falls back to the default XDG_RUNTIME_DIR location.
func FindSocketPath() string {
	// 1. Explicit override via environment variable
	if path := os.Getenv("CARTAGE_SOCKET"); path != "" {
		return path
	}

	// Build candidate paths
	var candidates []string

	// 2. Native host location (XDG_RUNTIME_DIR)
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		candidates = append(candidates, filepath.Join(xdg, "cartage.sock"))
	}

	// 3. Common container mount point
	candidates = append(candidates, containerSocketPath)

	// 4. Fallback location
	candidates = append(candidates, "/tmp/cartage.sock")

	// Return first existing socket
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Default to XDG path (daemon may not be running yet)
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		return filepath.Join(xdg, "cartage.sock")
	}

	return "/tmp/cartage.sock"
}
