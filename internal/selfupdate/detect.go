package selfupdate

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallMethod is how the running binary was installed, which determines the
// upgrade strategy.
type InstallMethod string

const (
	// MethodNPM: installed via `npm i -g linkai-cli`; upgrade with npm.
	MethodNPM InstallMethod = "npm"
	// MethodHomebrew: installed via Homebrew cask; upgrade with brew.
	MethodHomebrew InstallMethod = "homebrew"
	// MethodGo: installed via `go install`; upgrade with go install.
	MethodGo InstallMethod = "go"
	// MethodScript: installed via install.sh/install.ps1 or a manual download;
	// upgrade by re-running the install script.
	MethodScript InstallMethod = "script"
)

// DetectResult carries the detected method and the resolved binary path.
type DetectResult struct {
	Method InstallMethod
	Path   string // resolved absolute path of the running binary
}

// Detect inspects the running binary's path to infer how it was installed.
//
// Heuristics (checked in order):
//   - path contains "node_modules"                     → npm
//   - path is under a Homebrew Cellar/Caskroom or prefix → homebrew
//   - path is under GOPATH/bin or GOBIN                 → go
//   - otherwise                                          → script (install.sh)
func Detect() DetectResult {
	exe, err := os.Executable()
	if err != nil {
		return DetectResult{Method: MethodScript}
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return DetectResult{Method: methodFromPath(exe), Path: exe}
}

func methodFromPath(exe string) InstallMethod {
	lower := strings.ToLower(filepath.ToSlash(exe))

	// npm global installs live under a node_modules tree; the binary on PATH is
	// usually a symlink into it (already resolved above).
	if strings.Contains(lower, "/node_modules/") {
		return MethodNPM
	}

	// Homebrew: binaries resolve into the Cellar/Caskroom, or live under the
	// brew prefix (/opt/homebrew on Apple Silicon, /usr/local on Intel/Linux).
	if strings.Contains(lower, "/cellar/") || strings.Contains(lower, "/caskroom/") ||
		strings.Contains(lower, "/homebrew/") || strings.Contains(lower, "/linuxbrew/") {
		return MethodHomebrew
	}

	// go install places binaries in GOBIN or $GOPATH/bin (default ~/go/bin).
	if gobin := os.Getenv("GOBIN"); gobin != "" && underDir(exe, gobin) {
		return MethodGo
	}
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		for _, p := range filepath.SplitList(gopath) {
			if underDir(exe, filepath.Join(p, "bin")) {
				return MethodGo
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil && underDir(exe, filepath.Join(home, "go", "bin")) {
		return MethodGo
	}

	return MethodScript
}

func underDir(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// IsWindows reports whether the current OS is Windows (affects install command
// and terminal symbols).
func IsWindows() bool { return runtime.GOOS == "windows" }
