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
