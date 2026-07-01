// Package browser opens URLs in the user's default web browser.
//
// It is best-effort: on headless environments (containers, remote Linux
// sessions without a display, CI) opening will simply fail and the caller is
// expected to ignore the error and keep showing the URL for manual copy-paste.
package browser

import (
	"os/exec"
	"runtime"
)

// Open tries to open url in the default browser. It returns an error when no
// suitable opener is available or the command fails to start; callers on
// headless hosts should ignore it and fall back to printing the URL.
func Open(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		// rundll32 avoids cmd.exe quoting pitfalls with URL special chars (&, ?).
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		// Linux, BSD, etc. xdg-open is the freedesktop standard; absent on
		// headless boxes, in which case Start() returns an error we ignore.
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
