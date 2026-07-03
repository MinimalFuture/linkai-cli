package selfupdate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/MinimalFuture/linkai-cli/internal/config"
)

const (
	checkStateFile = "update-check.json"
	// checkInterval throttles background checks so we hit the network at most
	// once per interval, mirroring npm / gh behavior.
	checkInterval = 24 * time.Hour
)

// checkState is persisted to ~/.linkai/update-check.json between runs.
type checkState struct {
	LastCheck     int64  `json:"last_check"`     // Unix seconds
	LatestVersion string `json:"latest_version"` // last-seen latest version
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

// MaybeNotifyUpdate performs a throttled, best-effort background check and, when
// a newer version is available, returns the latest version string. It returns
// "" when there is nothing to report (up to date, throttled with no known
// update, offline, or explicitly disabled).
//
// Disable entirely with LINKAI_NO_UPDATE_NOTIFIER=1 (e.g. for CI). The check
// never blocks meaningfully: network calls are time-boxed and any error is
// swallowed.
func MaybeNotifyUpdate(current string) string {
	if os.Getenv("LINKAI_NO_UPDATE_NOTIFIER") == "1" {
		return ""
	}
	// "dev" builds are local; nagging about updates there is just noise.
	if ParseVersion(current) == nil {
		return ""
	}

	st := loadState()
	now := time.Now().Unix()

	// Within the throttle window: don't hit the network, but still surface a
	// previously discovered newer version.
	if now-st.LastCheck < int64(checkInterval.Seconds()) {
		if st.LatestVersion != "" && IsNewer(st.LatestVersion, current) {
			return Normalize(st.LatestVersion)
		}
		return ""
	}

	latest, err := FetchLatest(context.Background())
	if err != nil {
		return "" // offline / transient — try again after the interval
	}
	st.LastCheck = now
	st.LatestVersion = latest
	saveState(st)

	if IsNewer(latest, current) {
		return Normalize(latest)
	}
	return ""
}
