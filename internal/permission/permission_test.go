package permission

import (
	"strings"
	"testing"

	"github.com/MinimalFuture/linkai-cli/internal/auth"
)

func TestHas(t *testing.T) {
	tests := []struct {
		granted  string
		required Permission
		want     bool
	}{
		{"app:read user:read", AppRead, true},
		{"app:read user:read", ChatSend, false},
		{"", AppRead, false},
		{"knowledge:read knowledge:create knowledge:delete", KnowledgeDelete, true},
		{"app:read", AppRead, true},
		{"  app:read   user:read  ", UserRead, true},
	}
	for _, tt := range tests {
		got := Has(tt.granted, tt.required)
		if got != tt.want {
			t.Errorf("Has(%q, %q) = %v, want %v", tt.granted, tt.required, got, tt.want)
		}
	}
}

func TestCheck_NilToken(t *testing.T) {
	err := Check(nil, AppRead)
	if err == nil {
		t.Fatal("Check(nil, ...) = nil, want error")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("Check(nil) error = %v, want it to mention not logged in", err)
	}
}

func TestCheck_HasPermission(t *testing.T) {
	token := &auth.StoredToken{Scope: "app:read user:read"}
	if err := Check(token, AppRead); err != nil {
		t.Errorf("Check with valid permission: %v", err)
	}
}

func TestCheck_MissingPermission(t *testing.T) {
	token := &auth.StoredToken{Scope: "app:read"}
	err := Check(token, ChatSend)
	if err == nil {
		t.Fatal("Check with missing permission = nil, want error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "chat:send") {
		t.Errorf("error %q should mention required permission chat:send", msg)
	}
	if !strings.Contains(msg, "linkai auth login --scope") {
		t.Errorf("error %q should include remediation hint with --scope flag", msg)
	}
}

func TestCovered(t *testing.T) {
	tests := []struct {
		requested string
		granted   string
		want      bool
	}{
		{"app:read", "app:read user:read", true},
		{"app:read user:read", "app:read user:read chat:send", true},
		{"app:read chat:send", "app:read", false},
		{"", "app:read", true},
		{"app:read", "", false},
	}
	for _, tt := range tests {
		got := Covered(tt.requested, tt.granted)
		if got != tt.want {
			t.Errorf("Covered(%q, %q) = %v, want %v", tt.requested, tt.granted, got, tt.want)
		}
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		existing   string
		additional string
		want       string
	}{
		{"app:read", "user:read", "app:read user:read"},
		{"app:read user:read", "app:read", "app:read user:read"},
		{"", "app:read", "app:read"},
		{"app:read", "", "app:read"},
		{"app:read user:read", "chat:send db:read", "app:read user:read chat:send db:read"},
		{"app:read app:read", "user:read", "app:read user:read"},
	}
	for _, tt := range tests {
		got := Merge(tt.existing, tt.additional)
		if got != tt.want {
			t.Errorf("Merge(%q, %q) = %q, want %q", tt.existing, tt.additional, got, tt.want)
		}
	}
}

func TestDefaults_NoDeprecatedPermissions(t *testing.T) {
	d := Defaults()
	deprecated := []string{"app:write", "chat:read", "user:write", "model:read", "chat:write", "image:write", "video:write", "audio:write", "knowledge:write", "score:write"}
	fields := strings.Fields(d)
	set := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		set[f] = struct{}{}
	}
	for _, dep := range deprecated {
		if _, ok := set[dep]; ok {
			t.Errorf("Defaults() contains deprecated permission %q", dep)
		}
	}
}

func TestDefaults_IncludesEveryDeclaredPermission(t *testing.T) {
	d := Defaults()
	fields := strings.Fields(d)
	set := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		set[f] = struct{}{}
	}
	// Defaults grant read + create actions. The destructive edit/delete actions
	// and db:write are sensitive enough that we don't request them by default;
	// users opt in via the authorization page or --scope on login.
	expected := []Permission{
		AppRead, AppCreate, UserRead, ChatSend,
		KnowledgeRead, KnowledgeCreate, DBRead,
		ImageGen, VideoGen, AudioGen,
		PluginRead, PluginRun,
		WorkflowRead, WorkflowRun, WorkflowCreate,
		ScoreRead, ScoreBuy,
	}
	for _, p := range expected {
		if _, ok := set[p.String()]; !ok {
			t.Errorf("Defaults() missing %q", p)
		}
	}
}
