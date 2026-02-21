package open

import (
	"testing"
)

func TestMapPath(t *testing.T) {
	tests := []struct {
		name   string
		env    string // "" means unset
		unset  bool   // true means unset the env var entirely
		uri    string
		want   string
	}{
		{
			name:  "env unset",
			unset: true,
			uri:   "/workspace/file.txt",
			want:  "/workspace/file.txt",
		},
		{
			name: "env empty",
			env:  "",
			uri:  "/workspace/file.txt",
			want: "/workspace/file.txt",
		},
		{
			name:  "url passthrough",
			unset: true,
			uri:   "https://example.com",
			want:  "https://example.com",
		},
		{
			name: "url passthrough with env set",
			env:  "/workspace:/home/user/projects",
			uri:  "https://example.com/path",
			want: "https://example.com/path",
		},
		{
			name: "single mapping match",
			env:  "/workspace:/home/user/projects",
			uri:  "/workspace/my-project/file.pdf",
			want: "/home/user/projects/my-project/file.pdf",
		},
		{
			name: "single mapping no match",
			env:  "/workspace:/home/user/projects",
			uri:  "/other/path/file.txt",
			want: "/other/path/file.txt",
		},
		{
			name: "multiple mappings longest prefix wins",
			env:  "/workspace:/home/user/projects,/workspace/special:/mnt/special",
			uri:  "/workspace/special/file.txt",
			want: "/mnt/special/file.txt",
		},
		{
			name: "multiple mappings shorter match",
			env:  "/workspace:/home/user/projects,/workspace/special:/mnt/special",
			uri:  "/workspace/other/file.txt",
			want: "/home/user/projects/other/file.txt",
		},
		{
			name: "trailing slash in prefix",
			env:  "/workspace/:/home/user/projects/",
			uri:  "/workspace/file.txt",
			want: "/home/user/projects/file.txt",
		},
		{
			name: "path equals prefix exactly",
			env:  "/workspace:/home/user/projects",
			uri:  "/workspace",
			want: "/home/user/projects",
		},
		{
			name: "partial prefix should not match",
			env:  "/work:/home/user",
			uri:  "/workspace/file.txt",
			want: "/workspace/file.txt",
		},
		{
			name: "invalid entry without colon is skipped",
			env:  "invalid,/workspace:/home/user/projects",
			uri:  "/workspace/file.txt",
			want: "/home/user/projects/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.unset {
				t.Setenv("CARTAGE_PATH_MAP", "")
			} else {
				t.Setenv("CARTAGE_PATH_MAP", tt.env)
			}

			got := MapPath(tt.uri)
			if got != tt.want {
				t.Errorf("MapPath(%q) = %q, want %q", tt.uri, got, tt.want)
			}
		})
	}
}
