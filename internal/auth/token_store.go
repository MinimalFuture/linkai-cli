package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MinimalFuture/linkai-cli/internal/config"
	"github.com/MinimalFuture/linkai-cli/internal/keychain"
)

const (
	tokenFileName  = "token.json"
	refreshAheadMs = 5 * 60 * 1000 // 5 minutes
)

// StoredToken represents the full token set persisted to disk.
type StoredToken struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	ExpiresAt        int64  `json:"expires_at"`         // access token expiry, Unix ms
	RefreshExpiresAt int64  `json:"refresh_expires_at"` // refresh token expiry, Unix ms
	GrantedAt        int64  `json:"granted_at"`
}

func tokenPath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, tokenFileName), nil
}

// GetStoredToken reads the stored token. Tries the OS keychain first,
// then falls back to the file on disk.
func GetStoredToken() *StoredToken {
	// Try keychain first
	if keychain.Available() {
		if data, err := keychain.Load(); err == nil {
			var token StoredToken
			if json.Unmarshal(data, &token) == nil && token.AccessToken != "" {
				return &token
			}
		}
	}

	// Fallback to file
	p, err := tokenPath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var token StoredToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil
	}
	return &token
}

// SetStoredToken persists a token. Stores in the OS keychain if available,
// always writes a file as fallback.
func SetStoredToken(token *StoredToken) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	// Try keychain
	if keychain.Available() {
		if err := keychain.Store(data); err == nil {
			// Still write file for backward compatibility, but on keychain
			// success we don't fail if file write has issues.
			_ = writeTokenFile(data)
			return nil
		}
		// Keychain failed — fall through to file-only storage
	}

	return writeTokenFile(data)
}

// RemoveStoredToken removes the stored token from both keychain and disk.
func RemoveStoredToken() error {
	// Remove from keychain (best-effort)
	if keychain.Available() {
		_ = keychain.Remove()
	}

	// Remove from disk
	p, err := tokenPath()
	if err != nil {
		return err
	}
	err = os.Remove(p)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func writeTokenFile(data []byte) error {
	p, err := tokenPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}
	return os.WriteFile(p, data, 0600)
}

// TokenStatus determines the freshness of the access token.
// Returns "valid", "needs_refresh", or "expired".
func TokenStatus(token *StoredToken) string {
	now := time.Now().UnixMilli()
	if now < token.ExpiresAt-refreshAheadMs {
		return "valid"
	}
	if now < token.ExpiresAt {
		return "needs_refresh"
	}
	return "expired"
}

// MaskToken masks a token for safe logging.
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return "****" + token[len(token)-4:]
}
