package notify

import "encoding/json"

// Mode determines how the notification is displayed.
type Mode string

const (
	ModeToast   Mode = "toast"   // Non-blocking desktop notification
	ModeAlert   Mode = "alert"   // Blocking dialog with OK button
	ModeConfirm Mode = "confirm" // Blocking dialog with Yes/No buttons
)

// Payload is the action-specific data for a "notify" request.
type Payload struct {
	Title    string  `json:"title"`               // Required: notification title
	Body     *string `json:"body,omitempty"`      // Optional: notification body
	Mode     Mode    `json:"mode,omitempty"`      // Notification mode (defaults to toast)
	ToolHint *string `json:"tool_hint,omitempty"` // Preferred host tool (e.g. "kdialog", "zenity")
	Icon     *string `json:"icon,omitempty"`      // Freedesktop icon name or path
	IconData *string `json:"icon_data,omitempty"` // Base64-encoded image data
	IconType *string `json:"icon_type,omitempty"` // Image format (png, jpg, svg)
	Urgency  *string `json:"urgency,omitempty"`   // low, normal, critical
	Timeout  *uint32 `json:"timeout,omitempty"`   // Milliseconds (0 = never expire)
	Width    *uint32 `json:"width,omitempty"`     // Dialog width in pixels
	Height   *uint32 `json:"height,omitempty"`    // Dialog height in pixels
}

// UnmarshalJSON defaults mode to "toast" when omitted.
func (p *Payload) UnmarshalJSON(data []byte) error {
	type Alias Payload
	aux := &struct{ *Alias }{Alias: (*Alias)(p)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if p.Mode == "" {
		p.Mode = ModeToast
	}
	return nil
}

// Result is the action-specific data returned in Response.Data for notify actions.
type Result struct {
	ID        string `json:"id"`                  // UUID for this notification
	Confirmed *bool  `json:"confirmed,omitempty"` // Only set for confirm dialogs
}

// ExtractConfirmed extracts the "confirmed" field from a notify response's Data.
// Response.Data is an any that, after JSON round-trip, becomes map[string]any.
func ExtractConfirmed(data any) bool {
	m, ok := data.(map[string]any)
	if !ok {
		return false
	}
	confirmed, ok := m["confirmed"]
	if !ok {
		return false
	}
	b, ok := confirmed.(bool)
	return ok && b
}
