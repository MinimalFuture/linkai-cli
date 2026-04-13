package cmdutil

import (
	"fmt"
	"strings"

	"github.com/MinimalFuture/linkai-cli/internal/auth"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

// RequiredScopeKey is the cobra command annotation key for declaring required scope.
const RequiredScopeKey = "required_scope"

// DefaultReadScopes is the set of scopes granted by default on login.
const DefaultReadScopes = "app:read chat:read chat:write user:read workflow:read workflow:run knowledge:read db:read image:write video:write audio:write plugin:read plugin:run score:read score:write model:read"

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
// Returns a structured ExitError with a remediation hint when the check fails.
func CheckScope(token *auth.StoredToken, required string) error {
	if token == nil {
		return output.ErrAuth("not logged in: run 'linkai auth login'")
	}
	if HasScope(token.Scope, required) {
		return nil
	}
	merged := mergeScopes(token.Scope, required)
	return output.ErrWithHint(
		output.ExitAuth,
		fmt.Sprintf("permission denied: this operation requires the %q scope", required),
		fmt.Sprintf("linkai auth login --scope %q", merged),
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
