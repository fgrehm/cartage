package open

import (
	"path/filepath"
	"strings"
)

// ResolvePath resolves a relative file path to an absolute path.
// URLs (containing "://") and already-absolute paths are returned as-is.
// Symlinks are resolved when possible, falling back to simple absolute resolution.
func ResolvePath(arg string) string {
	if strings.Contains(arg, "://") || filepath.IsAbs(arg) {
		return arg
	}
	abs, err := filepath.Abs(arg)
	if err != nil {
		return arg
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}
