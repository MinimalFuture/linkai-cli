package auth

import (
	"testing"
	"time"
)

func TestTokenStatus_Valid(t *testing.T) {
	token := &StoredToken{
		ExpiresAt: time.Now().Add(10 * time.Minute).UnixMilli(),
	}
	if got := TokenStatus(token); got != "valid" {
		t.Errorf("TokenStatus = %q, want %q", got, "valid")
	}
}

func TestTokenStatus_NeedsRefresh(t *testing.T) {
	token := &StoredToken{
		ExpiresAt: time.Now().Add(3 * time.Minute).UnixMilli(),
	}
	if got := TokenStatus(token); got != "needs_refresh" {
		t.Errorf("TokenStatus = %q, want %q", got, "needs_refresh")
	}
}

func TestTokenStatus_Expired(t *testing.T) {
	token := &StoredToken{
		ExpiresAt: time.Now().Add(-1 * time.Minute).UnixMilli(),
	}
	if got := TokenStatus(token); got != "expired" {
		t.Errorf("TokenStatus = %q, want %q", got, "expired")
	}
}

func TestTokenStatus_BoundaryAtRefreshAhead(t *testing.T) {
	// Exactly at the refreshAheadMs boundary → should be needs_refresh
	token := &StoredToken{
		ExpiresAt: time.Now().Add(5 * time.Minute).UnixMilli(),
	}
	got := TokenStatus(token)
	if got != "needs_refresh" && got != "valid" {
		t.Errorf("TokenStatus at boundary = %q, want valid or needs_refresh", got)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "****"},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "****6789"},
		{"abcdefghijklmnop", "****mnop"},
	}
	for _, tt := range tests {
		if got := MaskToken(tt.input); got != tt.want {
			t.Errorf("MaskToken(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
