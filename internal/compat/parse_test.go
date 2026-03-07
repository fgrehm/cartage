package compat

import (
	"testing"

	"github.com/fgrehm/cartage/internal/notify"
)

func TestParseNotifySendArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(t *testing.T, p notify.Payload)
	}{
		{
			name:    "no title",
			args:    []string{"notify-send"},
			wantErr: true,
		},
		{
			name: "title only",
			args: []string{"notify-send", "Hello"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "Hello")
				assertNil(t, "Body", p.Body)
				assertEqual(t, "Mode", string(p.Mode), "toast")
				assertUint32Ptr(t, "Timeout", p.Timeout, 5000)
			},
		},
		{
			name: "title and body",
			args: []string{"notify-send", "Hello", "World"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "Hello")
				assertStringPtr(t, "Body", p.Body, "World")
			},
		},
		{
			name: "icon short flag",
			args: []string{"notify-send", "-i", "dialog-info", "Test"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "Icon", p.Icon, "dialog-info")
				assertEqual(t, "Title", p.Title, "Test")
			},
		},
		{
			name: "icon long flag",
			args: []string{"notify-send", "--icon", "dialog-info", "Test"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "Icon", p.Icon, "dialog-info")
			},
		},
		{
			name: "timeout",
			args: []string{"notify-send", "-t", "3000", "Test"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertUint32Ptr(t, "Timeout", p.Timeout, 3000)
			},
		},
		{
			name: "urgency",
			args: []string{"notify-send", "-u", "critical", "Test"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "Urgency", p.Urgency, "critical")
			},
		},
		{
			name: "ignored flags",
			args: []string{"notify-send", "-a", "MyApp", "-c", "email", "--hint", "int:x:1", "Test"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "Test")
				assertNil(t, "Body", p.Body)
			},
		},
		{
			name: "all flags combined",
			args: []string{"notify-send", "-i", "mail", "-t", "10000", "-u", "low", "Subject", "Body text"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "Subject")
				assertStringPtr(t, "Body", p.Body, "Body text")
				assertStringPtr(t, "Icon", p.Icon, "mail")
				assertUint32Ptr(t, "Timeout", p.Timeout, 10000)
				assertStringPtr(t, "Urgency", p.Urgency, "low")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parseNotifySendArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, p)
		})
	}
}

func TestParseDialogArgs(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		tool  string
		check func(t *testing.T, p notify.Payload)
	}{
		{
			name: "no args defaults to Dialog title",
			args: []string{"yad"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "Dialog")
				assertEqual(t, "Mode", string(p.Mode), "alert")
			},
		},
		{
			name: "title and text space-separated",
			args: []string{"yad", "--title", "My Title", "--text", "Body here"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "My Title")
				assertStringPtr(t, "Body", p.Body, "Body here")
			},
		},
		{
			name: "title and text equals-separated",
			args: []string{"zenity", "--title=My Title", "--text=Body here"},
			tool: "zenity",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "My Title")
				assertStringPtr(t, "Body", p.Body, "Body here")
				assertStringPtr(t, "ToolHint", p.ToolHint, "zenity")
			},
		},
		{
			name: "question mode",
			args: []string{"yad", "--question", "--text", "Continue?"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Mode", string(p.Mode), "confirm")
				assertEqual(t, "Title", p.Title, "Continue?")
			},
		},
		{
			name: "error urgency",
			args: []string{"zenity", "--error", "--text", "Failed"},
			tool: "zenity",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "Urgency", p.Urgency, "critical")
			},
		},
		{
			name: "info urgency",
			args: []string{"yad", "--info", "--text", "Done"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "Urgency", p.Urgency, "normal")
			},
		},
		{
			name: "warning urgency",
			args: []string{"yad", "--warning", "--text", "Careful"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "Urgency", p.Urgency, "normal")
			},
		},
		{
			name: "width and height space-separated",
			args: []string{"yad", "--width", "600", "--height", "300", "--text", "Big"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertUint32Ptr(t, "Width", p.Width, 600)
				assertUint32Ptr(t, "Height", p.Height, 300)
			},
		},
		{
			name: "width and height equals-separated",
			args: []string{"yad", "--width=600", "--height=300", "--text=Big"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertUint32Ptr(t, "Width", p.Width, 600)
				assertUint32Ptr(t, "Height", p.Height, 300)
			},
		},
		{
			name: "image flag",
			args: []string{"yad", "--image", "dialog-warning", "--text", "Warn"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "Icon", p.Icon, "dialog-warning")
			},
		},
		{
			name: "text fallback as title when no title",
			args: []string{"yad", "--text", "Just text"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "Just text")
				assertStringPtr(t, "Body", p.Body, "Just text")
			},
		},
		{
			name: "tool hint matches tool name",
			args: []string{"yad", "--text", "Test"},
			tool: "yad",
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "ToolHint", p.ToolHint, "yad")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parseDialogArgs(tt.args, tt.tool)
			tt.check(t, p)
		})
	}
}

func TestParseKdialogArgs(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, p notify.Payload)
	}{
		{
			name: "no args defaults to Dialog title",
			args: []string{"kdialog"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "Dialog")
				assertEqual(t, "Mode", string(p.Mode), "alert")
			},
		},
		{
			name: "passivepopup with timeout",
			args: []string{"kdialog", "--passivepopup", "Hello", "5"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Mode", string(p.Mode), "toast")
				assertEqual(t, "Title", p.Title, "Hello")
				assertStringPtr(t, "Body", p.Body, "Hello")
				assertUint32Ptr(t, "Timeout", p.Timeout, 5000)
			},
		},
		{
			name: "passivepopup with title",
			args: []string{"kdialog", "--passivepopup", "Body text", "3", "--title", "My Title"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Title", p.Title, "My Title")
				assertStringPtr(t, "Body", p.Body, "Body text")
				assertUint32Ptr(t, "Timeout", p.Timeout, 3000)
			},
		},
		{
			name: "msgbox",
			args: []string{"kdialog", "--msgbox", "Info message", "--title", "Info"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Mode", string(p.Mode), "alert")
				assertEqual(t, "Title", p.Title, "Info")
				assertStringPtr(t, "Urgency", p.Urgency, "normal")
			},
		},
		{
			name: "error dialog",
			args: []string{"kdialog", "--error", "Something broke"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Mode", string(p.Mode), "alert")
				assertStringPtr(t, "Urgency", p.Urgency, "critical")
				assertEqual(t, "Title", p.Title, "Something broke")
			},
		},
		{
			name: "yesno confirm",
			args: []string{"kdialog", "--yesno", "Are you sure?", "--title", "Confirm"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertEqual(t, "Mode", string(p.Mode), "confirm")
				assertEqual(t, "Title", p.Title, "Confirm")
				assertStringPtr(t, "Body", p.Body, "Are you sure?")
			},
		},
		{
			name: "icon flag",
			args: []string{"kdialog", "--msgbox", "Test", "--icon", "dialog-info"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "Icon", p.Icon, "dialog-info")
			},
		},
		{
			name: "tool hint is always kdialog",
			args: []string{"kdialog", "--msgbox", "Test"},
			check: func(t *testing.T, p notify.Payload) {
				t.Helper()
				assertStringPtr(t, "ToolHint", p.ToolHint, "kdialog")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parseKdialogArgs(tt.args)
			tt.check(t, p)
		})
	}
}

func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", field, got, want)
	}
}

func assertStringPtr(t *testing.T, field string, got *string, want string) {
	t.Helper()
	if got == nil {
		t.Errorf("%s is nil, want %q", field, want)
		return
	}
	if *got != want {
		t.Errorf("%s = %q, want %q", field, *got, want)
	}
}

func assertUint32Ptr(t *testing.T, field string, got *uint32, want uint32) {
	t.Helper()
	if got == nil {
		t.Errorf("%s is nil, want %d", field, want)
		return
	}
	if *got != want {
		t.Errorf("%s = %d, want %d", field, *got, want)
	}
}

func assertNil(t *testing.T, field string, got *string) {
	t.Helper()
	if got != nil {
		t.Errorf("%s = %q, want nil", field, *got)
	}
}
