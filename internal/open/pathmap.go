package open

import (
	"os"
	"strings"
)

// MapPath rewrites a URI using the CARTAGE_PATH_MAP environment variable.
// The env var contains comma-separated "container_prefix:host_prefix" pairs.
// Only paths (strings starting with "/") are rewritten; URLs pass through unchanged.
// When multiple mappings match, the longest container prefix wins.
// If CARTAGE_PATH_MAP is unset or no mapping matches, the URI is returned as-is.
func MapPath(uri string) string {
	if !strings.HasPrefix(uri, "/") {
		return uri
	}

	raw := os.Getenv("CARTAGE_PATH_MAP")
	if raw == "" {
		return uri
	}

	var bestFrom, bestTo string
	for entry := range strings.SplitSeq(raw, ",") {
		from, to, ok := strings.Cut(entry, ":")
		if !ok {
			continue
		}
		from = strings.TrimRight(from, "/")
		to = strings.TrimRight(to, "/")
		if from == "" {
			continue
		}

		if uri == from || strings.HasPrefix(uri, from+"/") {
			if len(from) > len(bestFrom) {
				bestFrom = from
				bestTo = to
			}
		}
	}

	if bestFrom == "" {
		return uri
	}

	return bestTo + uri[len(bestFrom):]
}
