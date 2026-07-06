package selfupdate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/MinimalFuture/linkai-cli/internal/config"
)

// gitDescribePattern matches the git-describe suffix that ldflags-less local
// builds carry (e.g. "v0.0.0-20260706071705-ee88ec1acc48+dirty" or
// "1.2.0-3-gabc1234"). Such builds are not published releases, so we never nag
// about updates for them.
var gitDescribePattern = regexp.MustCompile(`-\d+-g?[0-9a-f]{7,}|\+dirty|\bdevel\b`)

// isReleaseBuild reports whether version looks like a clean published release
// (e.g. "1.2.0" or "1.2.0-rc1") rather than a local/dev git-describe build.
func isReleaseBuild(version string) bool {
	if ParseVersion(version) == nil {
		return false
	}
	return !gitDescribePattern.MatchString(strings.TrimSpace(version))
}

const (
	checkStateFile = "update-check.json"
	// checkInterval throttles background network checks so we hit the network at
	// most once per interval, mirroring npm / gh behavior.
	checkInterval = 24 * time.Hour
	// notifyInterval throttles the on-screen notice so an outdated CLI nags at
	// most once per day, instead of on every command within the check window.
	notifyInterval = 24 * time.Hour
)

// checkState is persisted to ~/.linkai/update-check.json between runs.
type checkState struct {
	LastCheck     int64  `json:"last_check"`     // Unix seconds — last network check
	LatestVersion string `json:"latest_version"` // last-seen latest version
	LastNotified  int64  `json:"last_notified"`  // Unix seconds — last on-screen notice
}

func statePath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, checkStateFile), nil
}

func loadState() checkState {
	var st checkState
	p, err := statePath()
	if err != nil {
		return st
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return st
	}
	_ = json.Unmarshal(data, &st)
	return st
}

func saveState(st checkState) {
	p, err := statePath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return
	}
	data, err := json.Marshal(st)
	if err != nil {
		return
	}
	_ = os.WriteFile(p, data, 0600)
}

// disabled reports whether the update notifier is turned off, or the current
// build is a local/dev version for which nagging is just noise.
func disabled(current string) bool {
	if os.Getenv("LINKAI_NO_UPDATE_NOTIFIER") == "1" {
		return true
	}
	// Only nag published releases; skip "dev" and local git-describe builds.
	return !isReleaseBuild(current)
}

// RefreshLatestInBackground kicks off a throttled, best-effort network check for
// the latest version and caches the result to ~/.linkai/update-check.json. It
// returns immediately; the caller is expected to run it in a goroutine at
// startup. The next command run then reads the cache via NotifyIfDue without any
// network latency on the hot path.
//
// It hits the network at most once per checkInterval; within that window it is a
// no-op. All errors are swallowed (offline is not an error worth surfacing).
func RefreshLatestInBackground(current string) {
	if disabled(current) {
		return
	}
	st := loadState()
	now := time.Now().Unix()
	if now-st.LastCheck < int64(checkInterval.Seconds()) {
		return // cache still fresh
	}
	latest, err := FetchLatest(context.Background())
	if err != nil {
		return
	}
	st.LastCheck = now
	st.LatestVersion = latest
	saveState(st)
}

// NotifyIfDue reports the latest version to show in an update notice, or "" when
// nothing should be shown. It never touches the network — it only reads the
// cache populated by RefreshLatestInBackground — so it is safe to call on the
// command hot path.
//
// The notice is throttled to at most once per notifyInterval (once a day): when
// it returns a non-empty version, it records the notification time so repeated
// commands within the window stay quiet until the next day.
func NotifyIfDue(current string) string {
	if disabled(current) {
		return ""
	}
	st := loadState()
	if st.LatestVersion == "" || !IsNewer(st.LatestVersion, current) {
		return ""
	}
	now := time.Now().Unix()
	if now-st.LastNotified < int64(notifyInterval.Seconds()) {
		return "" // already nagged within the last day
	}
	st.LastNotified = now
	saveState(st)
	return Normalize(st.LatestVersion)
}
