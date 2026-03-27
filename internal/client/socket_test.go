package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindSocketPath_Explicit(t *testing.T) {
	t.Setenv("CARTAGE_SOCKET", "/custom/socket.sock")

	path := FindSocketPath()
	if path != "/custom/socket.sock" {
		t.Errorf("want /custom/socket.sock, got %s", path)
	}
}

func TestFindSocketPath_XDGRuntime(t *testing.T) {
	t.Setenv("CARTAGE_SOCKET", "")
	t.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")

	path := FindSocketPath()
	expected := "/run/user/1000/cartage.sock"
	if path != expected {
		t.Errorf("want %s, got %s", expected, path)
	}
}

func TestFindSocketPath_Fallback(t *testing.T) {
	t.Setenv("CARTAGE_SOCKET", "")
	t.Setenv("XDG_RUNTIME_DIR", "")

	path := FindSocketPath()
	if path != "/tmp/cartage.sock" {
		t.Errorf("want /tmp/cartage.sock, got %s", path)
	}
}

func TestFindSocketPath_ContainerMount(t *testing.T) {
	t.Setenv("CARTAGE_SOCKET", "")

	tmpDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", filepath.Join(tmpDir, "nonexistent"))

	// No sockets exist, should return the XDG path as default
	path := FindSocketPath()
	expected := filepath.Join(tmpDir, "nonexistent", "cartage.sock")
	if path != expected {
		t.Errorf("want %s, got %s", expected, path)
	}
}

func TestFindSocketPath_ExistingSocket(t *testing.T) {
	t.Setenv("CARTAGE_SOCKET", "")

	tmpDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", filepath.Join(tmpDir, "nonexistent"))

	// Create a fake socket at the container mount point
	containerPath := filepath.Join(tmpDir, "run-host")
	if err := os.MkdirAll(containerPath, 0o755); err != nil {
		t.Fatal(err)
	}
	sockPath := filepath.Join(containerPath, "cartage.sock")
	if err := os.WriteFile(sockPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	// Override the container path for testing
	origContainerPath := containerSocketPath
	containerSocketPath = sockPath
	defer func() { containerSocketPath = origContainerPath }()

	path := FindSocketPath()
	if path != sockPath {
		t.Errorf("want %s, got %s", sockPath, path)
	}
}
