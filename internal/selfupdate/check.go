package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// cdnLatestURL is the plain-text version pointer maintained by the release
	// pipeline (scripts/sync-cdn.sh). It holds a bare version like "0.1.0".
	cdnLatestURL = "https://cdn.link-ai.tech/cli/latest.txt"
	// githubLatestAPI is the fallback when the CDN is unreachable. It returns
	// the latest non-draft release as JSON.
	githubLatestAPI = "https://api.github.com/repos/MinimalFuture/linkai-cli/releases/latest"

	// RepoURL is the canonical repository, used to build release / changelog URLs.
	RepoURL = "https://github.com/MinimalFuture/linkai-cli"

	checkTimeout = 8 * time.Second
)

// FetchLatest returns the latest published version (bare, no leading "v"). It
// tries the CDN pointer first (fast, no rate limit) and falls back to the
// GitHub Releases API.
func FetchLatest(ctx context.Context) (string, error) {
	if v, err := fetchLatestFromCDN(ctx); err == nil && ParseVersion(v) != nil {
		return Normalize(v), nil
	}
	v, err := fetchLatestFromGitHub(ctx)
	if err != nil {
		return "", err
	}
	return Normalize(v), nil
}

func fetchLatestFromCDN(ctx context.Context) (string, error) {
	body, err := httpGet(ctx, cdnLatestURL, "")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func fetchLatestFromGitHub(ctx context.Context) (string, error) {
	body, err := httpGet(ctx, githubLatestAPI, "application/vnd.github+json")
	if err != nil {
		return "", err
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		return "", fmt.Errorf("parse GitHub release: %w", err)
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("GitHub release has no tag_name")
	}
	return rel.TagName, nil
}

func httpGet(ctx context.Context, url, accept string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "linkai-cli-update")
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

// ReleaseURL returns the GitHub release page for a version.
func ReleaseURL(version string) string {
	return RepoURL + "/releases/tag/" + Normalize(version)
}
