package cmdutil

import (
	"fmt"
	"strings"

	larkauth "github.com/yjr/linkai-cli/internal/auth"
)

// RequiredScopeKey is the cobra command annotation key for declaring required scope.
const RequiredScopeKey = "required_scope"

// DefaultReadScopes is the set of scopes granted by default on login.
const DefaultReadScopes = "app:read chat:read user:read workflow:read knowledge:read"

// HasScope reports whether tokenScope (space-separated) contains required.
func HasScope(tokenScope, required string) bool {
	for _, s := range strings.Fields(tokenScope) {
		if s == required {
			return true
		}
	}
	return false
}

// CheckScope verifies that token carries the required scope.
// Returns a descriptive error with a remediation hint when the check fails.
func CheckScope(token *larkauth.StoredToken, required string) error {
	if token == nil {
		return fmt.Errorf("not logged in: run 'linkai auth login'")
	}
	if HasScope(token.Scope, required) {
		return nil
	}
	return fmt.Errorf(
		"permission denied: this operation requires the %q scope\n"+
			"  Re-authorize to include it:\n"+
			"    linkai auth login --scope %q",
		required,
		mergeScopes(token.Scope, required),
	)
}

// mergeScopes returns a space-separated scope string that includes both
// the token's existing scopes and the new required scope (deduplicated).
func mergeScopes(existing, newScope string) string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strings.Fields(existing) {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	if !seen[newScope] {
		result = append(result, newScope)
	}
	return strings.Join(result, " ")
}
