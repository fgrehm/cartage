package secret

import (
	"fmt"
	"runtime"
	"sort"

	"github.com/godbus/dbus/v5"
)

// SecretRef identifies a secret in the keychain by its service and user.
type SecretRef struct {
	Service string `json:"service"`
	User    string `json:"user"`
}

// listSecrets returns all secrets in the default keychain collection.
//
// go-keyring has no List API, so this queries the Secret Service dbus interface
// directly. It is only supported on platforms that provide the Secret Service
// (Linux and some BSDs). On other platforms it returns a clear error.
func listSecrets() ([]SecretRef, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf(
			"secret list is not supported on %s (requires the Secret Service dbus interface)", runtime.GOOS)
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	service := conn.Object("org.freedesktop.secrets", "/org/freedesktop/secrets")

	var collectionPath dbus.ObjectPath
	if err := service.Call("org.freedesktop.Secret.Service.ReadAlias", 0, "default").Store(&collectionPath); err != nil {
		return nil, fmt.Errorf("failed to read default collection: %w", err)
	}
	if collectionPath == "" {
		return nil, fmt.Errorf("no default keyring collection found")
	}

	var items []dbus.ObjectPath
	itemsVariant, err := conn.Object("org.freedesktop.secrets", collectionPath).
		GetProperty("org.freedesktop.Secret.Collection.Items")
	if err != nil {
		return nil, fmt.Errorf("failed to list keyring items: %w", err)
	}
	if err := itemsVariant.Store(&items); err != nil {
		return nil, fmt.Errorf("failed to decode keyring items: %w", err)
	}

	refs := make([]SecretRef, 0, len(items))
	for _, itemPath := range items {
		attrsVariant, err := conn.Object("org.freedesktop.secrets", itemPath).
			GetProperty("org.freedesktop.Secret.Item.Attributes")
		if err != nil {
			continue
		}
		var attrs map[string]string
		if err := attrsVariant.Store(&attrs); err != nil {
			continue
		}
		refs = append(refs, SecretRef{
			Service: attrs["service"],
			User:    attrs["username"],
		})
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Service != refs[j].Service {
			return refs[i].Service < refs[j].Service
		}
		return refs[i].User < refs[j].User
	})

	return refs, nil
}
