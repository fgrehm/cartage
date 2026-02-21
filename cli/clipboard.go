package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fgrehm/cartage/internal/client"
	"github.com/fgrehm/cartage/internal/clipboard"
	"github.com/fgrehm/cartage/internal/protocol"
	"github.com/spf13/cobra"
)

var clipboardCmd = &cobra.Command{
	Use:   "clipboard",
	Short: "Read and write the host clipboard",
}

var clipboardCopyCmd = &cobra.Command{
	Use:   "copy [TEXT]",
	Short: "Copy text or image to the host clipboard",
	Long: `Copy text or image to the host clipboard.

Examples:
  echo "hello" | cartage clipboard copy
  cartage clipboard copy "hello world"
  cartage clipboard copy --image screenshot.png`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageFile, _ := cmd.Flags().GetString("image")
		imageType, _ := cmd.Flags().GetString("image-type")

		var payload clipboard.Payload
		payload.Op = clipboard.OpWrite

		if imageFile != "" {
			data, err := os.ReadFile(imageFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading image file: %v\n", err)
				os.Exit(1)
			}
			if imageType == "" {
				imageType = imageTypeFromExt(imageFile)
			}
			encoded := base64.StdEncoding.EncodeToString(data)
			payload.ImageData = &encoded
			payload.ImageType = &imageType
		} else {
			var text string
			if len(args) > 0 {
				text = args[0]
			} else {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
					os.Exit(1)
				}
				text = string(data)
			}
			payload.Text = &text
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		_, err = client.Send(protocol.Request{
			Version: protocol.CurrentVersion,
			Action:  "clipboard",
			Payload: payloadJSON,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var clipboardPasteCmd = &cobra.Command{
	Use:   "paste",
	Short: "Paste content from the host clipboard",
	Long: `Paste content from the host clipboard.

Examples:
  cartage clipboard paste
  cartage clipboard paste --output screenshot.png`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		outputFile, _ := cmd.Flags().GetString("output")

		payload := clipboard.Payload{Op: clipboard.OpRead}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		resp, err := client.Send(protocol.Request{
			Version: protocol.CurrentVersion,
			Action:  "clipboard",
			Payload: payloadJSON,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		result, err := clipboard.ParseResult(resp.Data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if outputFile != "" {
			switch result.ContentType {
			case clipboard.ContentText:
				if err := os.WriteFile(outputFile, []byte(result.Text), 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
					os.Exit(1)
				}
			case clipboard.ContentImage:
				data, err := base64.StdEncoding.DecodeString(result.ImageData)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error decoding image data: %v\n", err)
					os.Exit(1)
				}
				if err := os.WriteFile(outputFile, data, 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
					os.Exit(1)
				}
			}
			return
		}

		switch result.ContentType {
		case clipboard.ContentText:
			fmt.Print(result.Text)
		case clipboard.ContentImage:
			fmt.Fprintln(os.Stderr, "Error: clipboard contains an image; use --output FILE to save it")
			os.Exit(1)
		}
	},
}

func init() {
	clipboardCopyCmd.Flags().String("image", "", "Copy image file to clipboard")
	clipboardCopyCmd.Flags().String("image-type", "", `Override image MIME type (default: detected from extension or "png")`)
	clipboardPasteCmd.Flags().StringP("output", "o", "", "Write clipboard content to file")

	clipboardCmd.AddCommand(clipboardCopyCmd)
	clipboardCmd.AddCommand(clipboardPasteCmd)
}

func imageTypeFromExt(filename string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	switch ext {
	case "jpg", "jpeg":
		return "jpeg"
	case "gif":
		return "gif"
	case "webp":
		return "webp"
	default:
		return "png"
	}
}
