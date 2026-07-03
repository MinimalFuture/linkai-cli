package update

import (
	"fmt"

	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/selfupdate"
)

// changelogURL returns the release page for the target version.
func changelogURL(version string) string { return selfupdate.ReleaseURL(version) }

func reportCheck(opts *Options, cur, latest string, upToDate bool, detect selfupdate.DetectResult) error {
	io := opts.Factory.IOStreams
	if opts.JSON {
		action := "update_available"
		if upToDate {
			action = "up_to_date"
		}
		return output.PrintJSON(io.Out, map[string]interface{}{
			"ok":               true,
			"current_version":  cur,
			"latest_version":   latest,
			"update_available": !upToDate,
			"action":           action,
			"install_method":   string(detect.Method),
			"url":              selfupdate.ReleaseURL(latest),
		})
	}
	if upToDate {
		output.PrintSuccess(io.ErrOut, io.IsTerminal, fmt.Sprintf("linkai %s is up to date", cur))
		return nil
	}
	fmt.Fprintf(io.ErrOut, "Update available: %s -> %s\n", cur, latest)
	fmt.Fprintf(io.ErrOut, "  Release: %s\n", selfupdate.ReleaseURL(latest))
	fmt.Fprintf(io.ErrOut, "\nRun `linkai update` to install.\n")
	return nil
}

func reportUpToDate(opts *Options, cur, latest string) error {
	io := opts.Factory.IOStreams
	if opts.JSON {
		return output.PrintJSON(io.Out, map[string]interface{}{
			"ok":               true,
			"current_version":  cur,
			"latest_version":   latest,
			"update_available": false,
			"action":           "already_up_to_date",
		})
	}
	output.PrintSuccess(io.ErrOut, io.IsTerminal, fmt.Sprintf("linkai %s is already up to date", cur))
	return nil
}

func reportManualWindows(opts *Options, cur, latest string) error {
	io := opts.Factory.IOStreams
	if opts.JSON {
		return output.PrintJSON(io.Out, map[string]interface{}{
			"ok":              true,
			"current_version": cur,
			"latest_version":  latest,
			"action":          "manual_required",
			"message":         "automatic update is unavailable for script/manual installs on Windows",
			"url":             selfupdate.ReleaseURL(latest),
		})
	}
	fmt.Fprintf(io.ErrOut, "Automatic update is unavailable for a manual install on Windows.\n\n")
	fmt.Fprintf(io.ErrOut, "Re-run the installer in PowerShell:\n")
	fmt.Fprintf(io.ErrOut, "  irm https://cdn.link-ai.tech/cli/install.ps1 | iex\n\n")
	fmt.Fprintf(io.ErrOut, "Or download the release: %s\n", selfupdate.ReleaseURL(latest))
	return nil
}

func reportFailure(opts *Options, cur, latest string, detect selfupdate.DetectResult, res selfupdate.UpgradeResult) error {
	io := opts.Factory.IOStreams
	hint := permissionHint(detect.Method, res.Output)
	if opts.JSON {
		errObj := map[string]interface{}{
			"type":    "update_error",
			"message": fmt.Sprintf("update via %s failed: %v", detect.Method, res.Err),
		}
		if hint != "" {
			errObj["hint"] = hint
		}
		_ = output.PrintJSON(io.Out, map[string]interface{}{
			"ok":              false,
			"current_version": cur,
			"latest_version":  latest,
			"install_method":  string(detect.Method),
			"error":           errObj,
		})
		return output.ErrNetwork("update failed")
	}
	if hint != "" {
		return output.ErrWithHint(output.ExitGeneral, fmt.Sprintf("update via %s failed: %v", detect.Method, res.Err), hint)
	}
	return output.Errorf("update via %s failed: %v", detect.Method, res.Err)
}

func reportUpdated(opts *Options, cur, latest string, detect selfupdate.DetectResult, skillsErr error) error {
	io := opts.Factory.IOStreams
	if opts.JSON {
		out := map[string]interface{}{
			"ok":               true,
			"previous_version": cur,
			"current_version":  latest,
			"latest_version":   latest,
			"action":           "updated",
			"install_method":   string(detect.Method),
			"url":              selfupdate.ReleaseURL(latest),
		}
		if skillsErr != nil {
			out["skills_action"] = "failed"
			out["skills_warning"] = skillsErr.Error()
		} else {
			out["skills_action"] = "synced"
		}
		return output.PrintJSON(io.Out, out)
	}
	output.PrintSuccess(io.ErrOut, io.IsTerminal, fmt.Sprintf("updated linkai %s -> %s", cur, latest))
	fmt.Fprintf(io.ErrOut, "  Release: %s\n", changelogURL(latest))
	if skillsErr != nil {
		fmt.Fprintf(io.ErrOut, "  Note: skill sync failed: %v (run `linkai skill install`)\n", skillsErr)
	}
	return nil
}

func permissionHint(method selfupdate.InstallMethod, output string) string {
	if method == selfupdate.MethodNPM && containsAny(output, "EACCES", "permission denied") && !selfupdate.IsWindows() {
		return "npm lacks permission for the global prefix. Try `sudo linkai update`, or fix your npm prefix: https://docs.npmjs.com/resolving-eacces-permissions-errors"
	}
	return ""
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && indexInsensitive(s, sub) >= 0 {
			return true
		}
	}
	return false
}

// indexInsensitive is a tiny case-insensitive substring search to avoid pulling
// in strings.ToLower on potentially large command output twice.
func indexInsensitive(s, sub string) int {
	ls, lsub := len(s), len(sub)
	for i := 0; i+lsub <= ls; i++ {
		if equalFoldASCII(s[i:i+lsub], sub) {
			return i
		}
	}
	return -1
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
