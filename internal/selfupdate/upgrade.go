package selfupdate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// NpmPackage is the npm package name used for `npm i -g`.
const NpmPackage = "linkai-cli"

// UpgradeResult is the outcome of an upgrade attempt.
type UpgradeResult struct {
	Output string // combined stdout/stderr of the upgrade command
	Err    error
}

// Upgrade runs the upgrade command matching the detected install method,
// upgrading to the given version. The command inherits the process environment
// so package managers behave as they would when run directly.
func Upgrade(ctx context.Context, method InstallMethod, version string, io CommandIO) UpgradeResult {
	switch method {
	case MethodNPM:
		return runCommand(ctx, io, "npm", "install", "-g", fmt.Sprintf("%s@%s", NpmPackage, Normalize(version)))
	case MethodHomebrew:
		return runCommand(ctx, io, "brew", "upgrade", "--cask", "linkai")
	case MethodGo:
		return runCommand(ctx, io, "go", "install", fmt.Sprintf("github.com/MinimalFuture/linkai-cli@v%s", Normalize(version)))
	default:
		return upgradeViaScript(ctx, version, io)
	}
}

// upgradeViaScript re-runs the install script pinned to the target version.
// The script (install.sh) already handles download, checksum verification,
// atomic replacement and PATH — so we do not re-implement any of it here.
//
// On Windows there is no POSIX shell, so we return a manual instruction instead
// of trying to pipe into PowerShell from Go.
func upgradeViaScript(ctx context.Context, version string, io CommandIO) UpgradeResult {
	if IsWindows() {
		return UpgradeResult{Err: errManualWindows}
	}
	// LINKAI_NO_SKILL keeps the binary update focused; skills are refreshed
	// separately by the caller via `linkai skill install`. LINKAI_VERSION pins
	// the exact target so the update is deterministic.
	script := fmt.Sprintf(
		"curl -fsSL https://cdn.link-ai.tech/cli/install.sh | LINKAI_VERSION=%s LINKAI_NO_SKILL=1 sh",
		Normalize(version),
	)
	return runCommand(ctx, io, "sh", "-c", script)
}

var errManualWindows = fmt.Errorf("automatic script update is not supported on Windows")

// IsManualWindowsErr reports whether an upgrade error means the user must update
// manually on Windows.
func IsManualWindowsErr(err error) bool { return err == errManualWindows }

// CommandIO lets the caller stream command output to the terminal in human mode
// while capturing it in JSON mode.
type CommandIO struct {
	// Stream, when non-nil, receives the command's stdout+stderr live (in
	// addition to being captured in the returned Output).
	Stream io.Writer
}

func runCommand(ctx context.Context, cio CommandIO, name string, args ...string) UpgradeResult {
	path, err := exec.LookPath(name)
	if err != nil {
		return UpgradeResult{Err: fmt.Errorf("%q not found on PATH: %w", name, err)}
	}
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = os.Environ()

	var buf bytes.Buffer
	var sink io.Writer = &buf
	if cio.Stream != nil {
		sink = io.MultiWriter(&buf, cio.Stream)
	}
	cmd.Stdout = sink
	cmd.Stderr = sink

	err = cmd.Run()
	return UpgradeResult{Output: buf.String(), Err: err}
}
