package compat

import "testing"

func TestGetCompatModeSecretTool(t *testing.T) {
	tests := []struct {
		argv0 string
		want  string
	}{
		{"secret-tool", "secret-tool"},
		{"/usr/bin/secret-tool", "secret-tool"},
		{"secret-tool.sh", "secret-tool"},
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

func TestParseSecretToolArgsStore(t *testing.T) {
	sub, attrs, err := parseSecretToolArgs([]string{
		"secret-tool", "store", "--label=Password for 'alice' on 'myapp'",
		"service", "myapp", "username", "alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub != "store" {
		t.Errorf("subcommand: want store, got %s", sub)
	}
	if attrs["service"] != "myapp" {
		t.Errorf("service: want myapp, got %s", attrs["service"])
	}
	if attrs["username"] != "alice" {
		t.Errorf("username: want alice, got %s", attrs["username"])
	}
}

func TestParseSecretToolArgsLookup(t *testing.T) {
	sub, attrs, err := parseSecretToolArgs([]string{
		"secret-tool", "lookup", "service", "myapp", "username", "alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub != "lookup" {
		t.Errorf("subcommand: want lookup, got %s", sub)
	}
	if attrs["service"] != "myapp" {
		t.Errorf("service: want myapp, got %s", attrs["service"])
	}
	if attrs["username"] != "alice" {
		t.Errorf("username: want alice, got %s", attrs["username"])
	}
}

func TestParseSecretToolArgsMissingSubcommand(t *testing.T) {
	_, _, err := parseSecretToolArgs([]string{"secret-tool"})
	if err == nil {
		t.Error("expected error for missing subcommand")
	}
}

func TestParseSecretToolArgsMissingValue(t *testing.T) {
	_, _, err := parseSecretToolArgs([]string{"secret-tool", "lookup", "service"})
	if err == nil {
		t.Error("expected error for missing attribute value")
	}
}
