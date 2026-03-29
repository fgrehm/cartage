package notify

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// decodeAndSaveIcon decodes base64 icon data and saves it to a temp file.
// Returns the path to the temp file. Caller is responsible for cleanup.
func decodeAndSaveIcon(iconData string, iconType *string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(iconData)
	if err != nil {
		return "", fmt.Errorf("invalid base64: %w", err)
	}

	ext := "png"
	if iconType != nil {
		ext = *iconType
	}

	filename := fmt.Sprintf("cartage-%s.%s", uuid.New().String(), ext)
	tempPath := filepath.Join(os.TempDir(), filename)

	if err := os.WriteFile(tempPath, decoded, 0o644); err != nil {
		return "", fmt.Errorf("failed to write temp icon file: %w", err)
	}

	return tempPath, nil
}
