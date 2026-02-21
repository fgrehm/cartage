package protocol

import (
	"encoding/json"
	"testing"
)

func TestRequestMarshalRoundtrip(t *testing.T) {
	req := Request{
		Version: 1,
		Action:  "notify",
		Payload: json.RawMessage(`{"title":"hello"}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Request
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Version != 1 {
		t.Errorf("version: want 1, got %d", got.Version)
	}
	if got.Action != "notify" {
		t.Errorf("action: want notify, got %s", got.Action)
	}
	if string(got.Payload) != `{"title":"hello"}` {
		t.Errorf("payload: want {\"title\":\"hello\"}, got %s", got.Payload)
	}
}

func TestRequestMissingVersion(t *testing.T) {
	// JSON with no version field should unmarshal with version=0
	data := `{"action":"notify","payload":{}}`
	var req Request
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Version != 0 {
		t.Errorf("version: want 0, got %d", req.Version)
	}
}

func TestRequestEmptyPayload(t *testing.T) {
	data := `{"version":1,"action":"open"}`
	var req Request
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Action != "open" {
		t.Errorf("action: want open, got %s", req.Action)
	}
	if req.Payload != nil {
		t.Errorf("payload: want nil, got %s", req.Payload)
	}
}

func TestOkResponse(t *testing.T) {
	resp := OkResponse(map[string]bool{"confirmed": true})

	if resp.Status != "ok" {
		t.Errorf("status: want ok, got %s", resp.Status)
	}
	if resp.Error != "" {
		t.Errorf("error: want empty, got %s", resp.Error)
	}
	if resp.Data == nil {
		t.Fatal("data: want non-nil")
	}
}

func TestOkResponseNilData(t *testing.T) {
	resp := OkResponse(nil)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// data field should be omitted
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["data"]; ok {
		t.Errorf("data field should be omitted when nil")
	}
}

func TestErrorResponse(t *testing.T) {
	resp := ErrorResponse("something broke")

	if resp.Status != "error" {
		t.Errorf("status: want error, got %s", resp.Status)
	}
	if resp.Error != "something broke" {
		t.Errorf("error: want 'something broke', got %s", resp.Error)
	}
	if resp.Data != nil {
		t.Errorf("data: want nil, got %v", resp.Data)
	}
}

func TestResponseMarshalRoundtrip(t *testing.T) {
	resp := OkResponse(map[string]string{"id": "abc-123"})

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Response
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Status != "ok" {
		t.Errorf("status: want ok, got %s", got.Status)
	}
}
