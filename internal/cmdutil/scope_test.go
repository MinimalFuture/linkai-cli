package cmdutil

import (
	"testing"

	"github.com/MinimalFuture/linkai-cli/internal/auth"
)

func TestHasScope(t *testing.T) {
	tests := []struct {
		scopes   string
		required string
		want     bool
	}{
		{"app:read user:read", "app:read", true},
		{"app:read user:read", "app:write", false},
		{"", "app:read", false},
		{"knowledge:read knowledge:write knowledge:delete", "knowledge:delete", true},
		{"app:read", "app:read user:read", false},
	}
	for _, tt := range tests {
		got := HasScope(tt.scopes, tt.required)
		if got != tt.want {
			t.Errorf("HasScope(%q, %q) = %v, want %v", tt.scopes, tt.required, got, tt.want)
		}
	}
}

func TestCheckScope_Nil(t *testing.T) {
	if err := CheckScope(nil, "app:read"); err == nil {
		t.Error("CheckScope(nil, ...) = nil, want error")
	}
}

func TestCheckScope_HasScope(t *testing.T) {
	token := &auth.StoredToken{Scope: "app:read user:read"}
	if err := CheckScope(token, "app:read"); err != nil {
		t.Errorf("CheckScope with valid scope: %v", err)
	}
}

func TestCheckScope_MissingScope(t *testing.T) {
	token := &auth.StoredToken{Scope: "app:read"}
	err := CheckScope(token, "app:write")
	if err == nil {
		t.Error("CheckScope with missing scope = nil, want error")
	}
}
