package clipboard

import (
	"encoding/json"
	"testing"
)

func TestPayloadWriteText(t *testing.T) {
	text := "hello world"
	p := Payload{Op: OpWrite, Text: &text}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Payload
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Op != OpWrite {
		t.Errorf("op: want %s, got %s", OpWrite, got.Op)
	}
	if got.Text == nil || *got.Text != text {
		t.Errorf("text: want %q, got %v", text, got.Text)
	}
	if got.ImageData != nil {
		t.Errorf("image_data: want nil, got %v", got.ImageData)
	}
}

func TestPayloadWriteImage(t *testing.T) {
	imgData := "base64encodeddata"
	imgType := "png"
	p := Payload{Op: OpWrite, ImageData: &imgData, ImageType: &imgType}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Payload
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Op != OpWrite {
		t.Errorf("op: want %s, got %s", OpWrite, got.Op)
	}
	if got.ImageData == nil || *got.ImageData != imgData {
		t.Errorf("image_data: want %q, got %v", imgData, got.ImageData)
	}
	if got.ImageType == nil || *got.ImageType != imgType {
		t.Errorf("image_type: want %q, got %v", imgType, got.ImageType)
	}
	if got.Text != nil {
		t.Errorf("text: want nil, got %v", got.Text)
	}
}

func TestPayloadRead(t *testing.T) {
	p := Payload{Op: OpRead}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Payload
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Op != OpRead {
		t.Errorf("op: want %s, got %s", OpRead, got.Op)
	}
	if got.Text != nil || got.ImageData != nil {
		t.Errorf("read payload should have no text or image fields")
	}
}

func TestResultText(t *testing.T) {
	r := Result{ContentType: ContentText, Text: "hello"}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Result
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ContentType != ContentText {
		t.Errorf("content_type: want %s, got %s", ContentText, got.ContentType)
	}
	if got.Text != "hello" {
		t.Errorf("text: want hello, got %s", got.Text)
	}
	if got.ImageData != "" || got.ImageType != "" {
		t.Errorf("image fields should be empty for text result")
	}
}

func TestResultImage(t *testing.T) {
	r := Result{ContentType: ContentImage, ImageData: "abc123", ImageType: "png"}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Result
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ContentType != ContentImage {
		t.Errorf("content_type: want %s, got %s", ContentImage, got.ContentType)
	}
	if got.ImageData != "abc123" {
		t.Errorf("image_data: want abc123, got %s", got.ImageData)
	}
	if got.ImageType != "png" {
		t.Errorf("image_type: want png, got %s", got.ImageType)
	}
	if got.Text != "" {
		t.Errorf("text should be empty for image result")
	}
}
