package compat

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/fgrehm/cartage/internal/clipboard"
)

func TestWritePbpasteResult(t *testing.T) {
	imgBytes := []byte{0x89, 0x50, 0x4e, 0x47} // PNG magic bytes
	imgB64 := base64.StdEncoding.EncodeToString(imgBytes)

	tests := []struct {
		name       string
		result     clipboard.Result
		isTerminal bool
		wantOut    string
		wantBytes  []byte
		wantErr    bool
	}{
		{
			name: "text content",
			result: clipboard.Result{
				ContentType: clipboard.ContentText,
				Text:        "hello world",
			},
			isTerminal: true,
			wantOut:    "hello world",
		},
		{
			name: "text content piped",
			result: clipboard.Result{
				ContentType: clipboard.ContentText,
				Text:        "hello world",
			},
			isTerminal: false,
			wantOut:    "hello world",
		},
		{
			name: "image content on terminal",
			result: clipboard.Result{
				ContentType: clipboard.ContentImage,
				ImageData:   imgB64,
				ImageType:   "png",
			},
			isTerminal: true,
			wantErr:    true,
		},
		{
			name: "image content piped",
			result: clipboard.Result{
				ContentType: clipboard.ContentImage,
				ImageData:   imgB64,
				ImageType:   "png",
			},
			isTerminal: false,
			wantBytes:  imgBytes,
		},
		{
			name: "image content invalid base64",
			result: clipboard.Result{
				ContentType: clipboard.ContentImage,
				ImageData:   "not-valid-base64!@#$",
			},
			isTerminal: false,
			wantErr:    true,
		},
		{
			name: "empty text content",
			result: clipboard.Result{
				ContentType: clipboard.ContentText,
				Text:        "",
			},
			isTerminal: true,
			wantOut:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writePbpasteResult(tt.result, tt.isTerminal, &buf)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantBytes != nil {
				if !bytes.Equal(buf.Bytes(), tt.wantBytes) {
					t.Errorf("output bytes = %v, want %v", buf.Bytes(), tt.wantBytes)
				}
			} else if buf.String() != tt.wantOut {
				t.Errorf("output = %q, want %q", buf.String(), tt.wantOut)
			}
		})
	}
}
