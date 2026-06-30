// Package permission centralizes the CLI's authorization model.
//
// A Permission is a typed string of the form "resource:action" (e.g. "app:read").
// Commands declare what they need via the cobra annotation key RequiredKey, and
// the root command checks the current token before running. Wire-format fields
// (StoredToken.Scope, the --scope flag, server parameters) keep the OAuth term
// "scope"; internal Go code uses "permission".
package permission

import (
	"fmt"
	"strings"

	"github.com/MinimalFuture/linkai-cli/internal/auth"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

// Permission is a single authorization unit, e.g. "app:read".
type Permission string

func (p Permission) String() string { return string(p) }

// All permissions the CLI declares. Action verbs match command semantics
// (chat:send, image:gen, score:buy) instead of an overloaded :write.
const (
	AppRead         Permission = "app:read"
	AppCreate       Permission = "app:create"
	AppUpdate       Permission = "app:update"
	AppDelete       Permission = "app:delete"
	UserRead        Permission = "user:read"
	ChatSend        Permission = "chat:send"
	KnowledgeRead   Permission = "knowledge:read"
	KnowledgeCreate Permission = "knowledge:create"
	KnowledgeUpdate Permission = "knowledge:update"
	KnowledgeDelete Permission = "knowledge:delete"
	DBRead          Permission = "db:read"
	DBWrite         Permission = "db:write"
	ImageGen        Permission = "image:gen"
	VideoGen        Permission = "video:gen"
	AudioGen        Permission = "audio:gen"
	PluginRead      Permission = "plugin:read"
	PluginRun       Permission = "plugin:run"
	WorkflowRead    Permission = "workflow:read"
	WorkflowRun     Permission = "workflow:run"
	ScoreRead       Permission = "score:read"
	ScoreBuy        Permission = "score:buy"
)

// RequiredKey is the cobra command annotation key used to declare the
// permission a command needs before it runs.
const RequiredKey = "required_permission"

// defaults is the permission set requested at login by default.
// Only includes permissions actually used by the CLI.
var defaults = []Permission{
	AppRead,
	UserRead,
	ChatSend,
	KnowledgeRead,
	DBRead,
	ImageGen,
	VideoGen,
	AudioGen,
	PluginRead,
	PluginRun,
	WorkflowRead,
	WorkflowRun,
	ScoreRead,
	ScoreBuy,
}

// Defaults returns the space-separated permission string requested on login.
func Defaults() string {
	parts := make([]string, len(defaults))
	for i, p := range defaults {
		parts[i] = p.String()
	}
	return strings.Join(parts, " ")
}

// Has reports whether granted (space-separated) contains required.
func Has(granted string, required Permission) bool {
	r := required.String()
	for _, s := range strings.Fields(granted) {
		if s == r {
			return true
		}
	}
	return false
}

// Check verifies token carries required and returns a structured ExitError
// with a remediation hint when it does not.
func Check(token *auth.StoredToken, required Permission) error {
	if token == nil {
		return output.ErrAuth("not logged in: run 'linkai auth login'")
	}
	if Has(token.Scope, required) {
		return nil
	}
	merged := Merge(token.Scope, required.String())
	return output.ErrWithHint(
		output.ExitAuth,
		fmt.Sprintf("permission denied: this operation requires the %q permission", required),
		fmt.Sprintf("linkai auth login --scope %q", merged),
	)
}

// Covered reports whether granted (space-separated) contains every entry
// in requested (space-separated).
func Covered(requested, granted string) bool {
	grantedSet := make(map[string]struct{})
	for _, s := range strings.Fields(granted) {
		grantedSet[s] = struct{}{}
	}
	for _, s := range strings.Fields(requested) {
		if _, ok := grantedSet[s]; !ok {
			return false
		}
	}
	return true
}

// Merge returns a deduplicated space-separated string containing every
// entry in existing plus additional.
func Merge(existing, additional string) string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strings.Fields(existing) {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range strings.Fields(additional) {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return strings.Join(result, " ")
}
