package notify

import (
	"encoding/json"
	"testing"
)

func TestPayloadMinimal(t *testing.T) {
	data := `{"title":"Test"}`
	var p Payload
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if p.Title != "Test" {
		t.Errorf("title: want Test, got %s", p.Title)
	}
	if p.Mode != ModeToast {
		t.Errorf("mode: want toast, got %s", p.Mode)
	}
	if p.Body != nil {
		t.Errorf("body: want nil, got %v", p.Body)
	}
}

func TestPayloadFull(t *testing.T) {
	body := "Test body"
	icon := "dialog-warning"
	urgency := "critical"
	timeout := uint32(5000)

	p := Payload{
		Title:   "Test",
		Body:    &body,
		Mode:    ModeAlert,
		Icon:    &icon,
		Urgency: &urgency,
		Timeout: &timeout,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Payload
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Title != p.Title {
		t.Errorf("title mismatch")
	}
	if *got.Body != *p.Body {
		t.Errorf("body mismatch")
	}
	if got.Mode != p.Mode {
		t.Errorf("mode mismatch")
	}
	if *got.Icon != *p.Icon {
		t.Errorf("icon mismatch")
	}
}

func TestPayloadModeDefaults(t *testing.T) {
	data := `{"title":"Test","body":"Body"}`
	var p Payload
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Mode != ModeToast {
		t.Errorf("mode: want toast, got %s", p.Mode)
	}
}

func TestResultSuccess(t *testing.T) {
	r := Result{ID: "test-uuid-123"}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Result
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != "test-uuid-123" {
		t.Errorf("id: want test-uuid-123, got %s", got.ID)
	}
	if got.Confirmed != nil {
		t.Errorf("confirmed: want nil, got %v", got.Confirmed)
	}
}

func TestResultConfirm(t *testing.T) {
	confirmed := true
	r := Result{ID: "test-uuid", Confirmed: &confirmed}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Result
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Confirmed == nil || *got.Confirmed != true {
		t.Error("confirmed not preserved through serialization")
	}
}
