// Package keychain provides secure token storage using the OS keychain.
// Falls back to file-based storage when the keychain is unavailable.
package keychain

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	serviceName = "linkai-cli"
	accountName = "token"
)

// Store saves data to the OS keychain. Returns an error if the keychain
// is not available, allowing callers to fall back to file storage.
func Store(data []byte) error {
	switch runtime.GOOS {
	case "darwin":
		return macosStore(data)
	default:
		return fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
}

// Load reads data from the OS keychain. Returns an error if the keychain
// is not available or the entry doesn't exist.
func Load() ([]byte, error) {
	switch runtime.GOOS {
	case "darwin":
		return macosLoad()
	default:
		return nil, fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
}

// Remove deletes the entry from the OS keychain.
func Remove() error {
	switch runtime.GOOS {
	case "darwin":
		return macosRemove()
	default:
		return nil
	}
}

// Available reports whether keychain storage is supported on this platform.
func Available() bool {
	switch runtime.GOOS {
	case "darwin":
		return true
	default:
		return false
	}
}

// ── macOS Keychain (via security command) ──

func macosStore(data []byte) error {
	// Delete existing entry first (ignore error if not found)
	_ = macosRemove()

	// JSON-encode data for safe storage
	encoded, err := json.Marshal(string(data))
	if err != nil {
		return err
	}

	cmd := exec.Command("security", "add-generic-password",
		"-s", serviceName,
		"-a", accountName,
		"-w", string(encoded),
		"-U", // update if exists
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("keychain store failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func macosLoad() ([]byte, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", serviceName,
		"-a", accountName,
		"-w", // output password only
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("keychain load failed: %w", err)
	}

	// Unescape JSON encoding
	raw := strings.TrimSpace(string(out))
	var decoded string
	if jsonErr := json.Unmarshal([]byte(raw), &decoded); jsonErr != nil {
		// Not JSON-encoded; return raw value as a fallback for legacy entries.
		return []byte(raw), nil //nolint:nilerr // intentional fallback when stored value isn't JSON-encoded
	}
	return []byte(decoded), nil
}

func macosRemove() error {
	cmd := exec.Command("security", "delete-generic-password",
		"-s", serviceName,
		"-a", accountName,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		errStr := strings.TrimSpace(string(out))
		if strings.Contains(errStr, "could not be found") {
			return nil // not found is OK
		}
		return fmt.Errorf("keychain remove failed: %s: %w", errStr, err)
	}
	return nil
}
