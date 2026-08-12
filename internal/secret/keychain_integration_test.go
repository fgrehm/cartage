package secret

import (
	"os"
	"os/exec"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestKeychainRoundTrip is the Phase 0 tracer bullet. It validates the riskiest
// assumption: that go-keyring can store and retrieve a secret in the real host
// keychain from this environment. It uses the real keychain (not MockInit)
// because the point is to prove the integration, not the handler logic.
//
// It skips when the keychain is unreachable (e.g. CI without a dbus session
// bus), so the suite stays green there while still proving the integration
// locally.
func TestKeychainRoundTrip(t *testing.T) {
	if !keychainAvailable() {
		t.Skip("host keychain unavailable (no dbus session bus or secret-tool); skipping integration test")
	}

	service := "cartage-test"
	user := "tracer-bullet"
	secret := "s3cr3t-value"

	if err := keyring.Set(service, user, secret); err != nil {
		t.Fatalf("keyring.Set failed: %v", err)
	}
	t.Cleanup(func() {
		_ = keyring.Delete(service, user)
	})

	got, err := keyring.Get(service, user)
	if err != nil {
		t.Fatalf("keyring.Get failed: %v", err)
	}

	if got != secret {
		t.Errorf("round-trip mismatch: want %q, got %q", secret, got)
	}
}

// keychainAvailable reports whether the host keychain is reachable: a dbus
// session bus must be set and the secret-tool binary must be present.
func keychainAvailable() bool {
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		return false
	}
	if _, err := exec.LookPath("secret-tool"); err != nil {
		return false
	}
	return true
}
