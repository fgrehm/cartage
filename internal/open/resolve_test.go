package open

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath(t *testing.T) {
	t.Run("URL passthrough", func(t *testing.T) {
		got := ResolvePath("https://example.com/path")
		if got != "https://example.com/path" {
			t.Errorf("ResolvePath(URL) = %q, want unchanged", got)
		}
	})

	t.Run("absolute path passthrough", func(t *testing.T) {
		got := ResolvePath("/absolute/path/file.pdf")
		if got != "/absolute/path/file.pdf" {
			t.Errorf("ResolvePath(absolute) = %q, want unchanged", got)
		}
	})

	t.Run("relative path becomes absolute", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "file.pdf")
		if err := os.WriteFile(f, nil, 0o644); err != nil {
			t.Fatal(err)
		}

		t.Chdir(dir)

		got := ResolvePath("file.pdf")
		if got != f {
			t.Errorf("ResolvePath(relative) = %q, want %q", got, f)
		}
	})

	t.Run("relative path with subdirectory", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub", "dir")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		f := filepath.Join(sub, "file.pdf")
		if err := os.WriteFile(f, nil, 0o644); err != nil {
			t.Fatal(err)
		}

		t.Chdir(dir)

		got := ResolvePath("sub/dir/file.pdf")
		if got != f {
			t.Errorf("ResolvePath(sub/dir/file.pdf) = %q, want %q", got, f)
		}
	})

	t.Run("symlink resolved to real path", func(t *testing.T) {
		dir := t.TempDir()
		real := filepath.Join(dir, "real.pdf")
		if err := os.WriteFile(real, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(dir, "link.pdf")
		if err := os.Symlink(real, link); err != nil {
			t.Fatal(err)
		}

		t.Chdir(dir)

		got := ResolvePath("link.pdf")
		if got != real {
			t.Errorf("ResolvePath(symlink) = %q, want %q", got, real)
		}
	})

	t.Run("nonexistent relative path falls back to Abs", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		got := ResolvePath("nonexistent.pdf")
		want := filepath.Join(dir, "nonexistent.pdf")
		if got != want {
			t.Errorf("ResolvePath(nonexistent) = %q, want %q", got, want)
		}
	})
}
