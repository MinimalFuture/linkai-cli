package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/MinimalFuture/linkai-cli/internal/validate"
)

// PrintJSON writes v as indented JSON to w.
func PrintJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// PrintTable writes a tab-aligned table to w.
// headers are printed as the first row in upper-case.
func PrintTable(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range rows {
		sanitized := make([]string, len(row))
		for i, cell := range row {
			sanitized[i] = validate.SanitizeOutput(cell)
		}
		fmt.Fprintln(tw, strings.Join(sanitized, "\t"))
	}
	tw.Flush()
}

// linkLabels maps server-returned link keys to human labels, in display order.
var linkLabels = []struct{ key, label string }{
	{"config", "Configure"},
	{"console", "Open in console"},
	{"chat", "Chat"},
}

// PrintLinks renders a server-provided `links` object (map of name → url).
// The server owns URL construction so the CLI never hard-codes domains; this
// helper just prints whatever links it receives, in a stable order. It is a
// no-op when links is nil/empty or not a map.
func PrintLinks(w io.Writer, links interface{}) {
	m, ok := links.(map[string]interface{})
	if !ok || len(m) == 0 {
		return
	}
	printed := make(map[string]bool, len(m))
	emit := func(key string) {
		if url, ok := m[key].(string); ok && url != "" {
			label := key
			for _, l := range linkLabels {
				if l.key == key {
					label = l.label
					break
				}
			}
			fmt.Fprintf(w, "  %s: %s\n", label, url)
			printed[key] = true
		}
	}
	for _, l := range linkLabels {
		emit(l.key)
	}
	// Print any remaining (unknown) keys so new server links still surface.
	for key := range m {
		if !printed[key] {
			emit(key)
		}
	}
}

// PrintSuccess writes a success message. Green when terminal supports ANSI.
func PrintSuccess(w io.Writer, isTerminal bool, msg string) {
	if isTerminal {
		fmt.Fprintf(w, "\033[32m✓\033[0m %s\n", msg)
	} else {
		fmt.Fprintf(w, "✓ %s\n", msg)
	}
}

// PrintError writes an error message. Red when terminal supports ANSI.
func PrintError(w io.Writer, isTerminal bool, msg string) {
	if isTerminal {
		fmt.Fprintf(w, "\033[31m✗\033[0m %s\n", msg)
	} else {
		fmt.Fprintf(w, "✗ %s\n", msg)
	}
}
