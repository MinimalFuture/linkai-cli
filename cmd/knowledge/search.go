package knowledge

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type SearchOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	DryRun  bool
	Code    string
	Query   string
	K       int
	Mode    string
}

type SearchResult struct {
	List              []SearchHit `json:"list"`
	KeywordSearchList []SearchHit `json:"keyword_search_list,omitempty"`
}

type SearchHit struct {
	Text        string  `json:"text"`
	Question    string  `json:"question"`
	Answer      string  `json:"answer"`
	Similarity  float64 `json:"similarity"`
	RerankScore float64 `json:"rerank_score"`
	Source      string  `json:"source"`
	DataType    string  `json:"data_type"`
}

func NewCmdKnowledgeSearch(f *cmdutil.Factory, runF func(*SearchOptions) error) *cobra.Command {
	opts := &SearchOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "search <code> <query>",
		Short: "Search a knowledge base",
		Args:  cobra.ExactArgs(2),
		Annotations: map[string]string{
			permission.RequiredKey: permission.KnowledgeRead.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			opts.Query = args[1]
			if runF != nil {
				return runF(opts)
			}
			return searchRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().IntVar(&opts.K, "k", 5, "number of results to return")
	cmd.Flags().StringVar(&opts.Mode, "mode", "", "search mode: vector, keyword, or hybrid (default hybrid)")

	return cmd
}

// searchModes maps user-friendly flag values (and the backend enum values
// themselves) to the enum expected by the backend. When --mode is omitted we
// default to hybrid search.
var searchModes = map[string]string{
	"":               "HYBRID_SEARCH",
	"hybrid":         "HYBRID_SEARCH",
	"vector":         "VECTOR_SEARCH",
	"keyword":        "KEYWORD_SEARCH",
	"hybrid_search":  "HYBRID_SEARCH",
	"vector_search":  "VECTOR_SEARCH",
	"keyword_search": "KEYWORD_SEARCH",
}

func resolveSearchMode(mode string) (string, error) {
	if m, ok := searchModes[strings.ToLower(strings.TrimSpace(mode))]; ok {
		return m, nil
	}
	return "", fmt.Errorf("invalid --mode %q: expected one of vector, keyword, hybrid", mode)
}

func searchRun(opts *SearchOptions) error {
	mode, err := resolveSearchMode(opts.Mode)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"code":       opts.Code,
		"query":      opts.Query,
		"k":          opts.K,
		"searchMode": mode,
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/knowledge/search",
			Body:   body,
		})
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Post(opts.Ctx, "/cli/knowledge/search", body)
	if err != nil {
		return fmt.Errorf("failed to search knowledge base: %w", err)
	}

	var result SearchResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	if len(result.List) == 0 && len(result.KeywordSearchList) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No results found.")
		return nil
	}

	out := opts.Factory.IOStreams.Out
	// When hybrid search returns keyword hits, label the vector section so the
	// two result sets are distinguishable.
	if len(result.KeywordSearchList) > 0 {
		fmt.Fprintln(out, "== Vector search ==")
	}
	printSearchHits(out, result.List)

	if len(result.KeywordSearchList) > 0 {
		fmt.Fprintln(out, "== Keyword search ==")
		printSearchHits(out, result.KeywordSearchList)
	}

	return nil
}

func printSearchHits(out io.Writer, hits []SearchHit) {
	for i, hit := range hits {
		scoreStr := fmt.Sprintf("similarity=%.4f", hit.Similarity)
		if hit.RerankScore > 0 {
			scoreStr += fmt.Sprintf("  rerank=%.4f", hit.RerankScore)
		}
		fmt.Fprintf(out, "[%d] %s  type=%s  source=%s\n", i+1, scoreStr, hit.DataType, hit.Source)
		if hit.DataType == "QA" {
			fmt.Fprintf(out, "    Q: %s\n", truncateRunes(hit.Question, 200))
			fmt.Fprintf(out, "    A: %s\n\n", truncateRunes(hit.Answer, 200))
		} else {
			fmt.Fprintf(out, "    %s\n\n", truncateRunes(hit.Text, 200))
		}
	}
}

func truncateRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
