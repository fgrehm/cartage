package compat

import "testing"

func TestGetCompatMode(t *testing.T) {
	tests := []struct {
		argv0 string
		want  string
	}{
		{"notify-send", "notify-send"},
		{"/usr/bin/notify-send", "notify-send"},
		{"notify-send.sh", "notify-send"},
		{"yad", "yad"},
		{"/usr/local/bin/yad", "yad"},
		{"zenity", "zenity"},
		{"/snap/bin/zenity", "zenity"},
		{"kdialog", "kdialog"},
		{"/usr/bin/kdialog", "kdialog"},
		{"xdg-open", "xdg-open"},
		{"/usr/bin/xdg-open", "xdg-open"},
		{"pbcopy", "pbcopy"},
		{"/usr/local/bin/pbcopy", "pbcopy"},
		{"pbpaste", "pbpaste"},
		{"/usr/local/bin/pbpaste", "pbpaste"},
		{"cartage", ""},
		{"/usr/local/bin/cartage", ""},
		{"something-else", ""},
	}

	for _, tt := range tests {
		t.Run(tt.argv0, func(t *testing.T) {
			got := GetCompatMode(tt.argv0)
			if got != tt.want {
				t.Errorf("GetCompatMode(%q) = %q, want %q", tt.argv0, got, tt.want)
			}
		})
	}
}

func TestIsCompatMode(t *testing.T) {
	if !IsCompatMode("notify-send") {
		t.Error("notify-send should be compat mode")
	}
	if !IsCompatMode("xdg-open") {
		t.Error("xdg-open should be compat mode")
	}
	if !IsCompatMode("pbcopy") {
		t.Error("pbcopy should be compat mode")
	}
	if !IsCompatMode("pbpaste") {
		t.Error("pbpaste should be compat mode")
	}
	if IsCompatMode("cartage") {
		t.Error("cartage should not be compat mode")
	}
}
