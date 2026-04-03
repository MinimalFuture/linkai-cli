package knowledge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yjr/linkai-cli/internal/cmdutil"
	"github.com/yjr/linkai-cli/internal/output"
)

type SearchOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	Code    string
	Query   string
	K       int
}

type SearchResult struct {
	List []SearchHit `json:"list"`
}

type SearchHit struct {
	Text       string  `json:"text"`
	Similarity float64 `json:"similarity"`
	FileName   string  `json:"fileName"`
	FileID     string  `json:"fileId"`
}

func NewCmdKnowledgeSearch(f *cmdutil.Factory, runF func(*SearchOptions) error) *cobra.Command {
	opts := &SearchOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "search <code> <query>",
		Short: "Search a knowledge base",
		Args:  cobra.ExactArgs(2),
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "knowledge:read",
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
	cmd.Flags().IntVar(&opts.K, "k", 5, "number of results to return")

	return cmd
}

func searchRun(opts *SearchOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Post(opts.Ctx, "/api/cli/knowledge/search", map[string]interface{}{
		"code":  opts.Code,
		"query": opts.Query,
		"k":     opts.K,
	})
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

	if len(result.List) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No results found.")
		return nil
	}

	for i, hit := range result.List {
		fmt.Fprintf(opts.Factory.IOStreams.Out, "[%d] similarity=%.4f  file=%s\n", i+1, hit.Similarity, hit.FileName)
		fmt.Fprintf(opts.Factory.IOStreams.Out, "    %s\n\n", truncateRunes(hit.Text, 200))
	}

	return nil
}

func truncateRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
