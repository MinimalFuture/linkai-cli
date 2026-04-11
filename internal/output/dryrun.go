package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// DryRunInfo describes the API request that would be sent.
type DryRunInfo struct {
	Method string      `json:"method"`
	URL    string      `json:"url"`
	Params interface{} `json:"params,omitempty"`
	Body   interface{} `json:"body,omitempty"`
}

// PrintDryRun outputs a dry-run summary without executing the request.
func PrintDryRun(w io.Writer, info DryRunInfo) error {
	fmt.Fprintln(w, "=== Dry Run ===")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(info)
}
