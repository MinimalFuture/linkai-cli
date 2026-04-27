package knowledge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type ListOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
}

type KnowledgeBase struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type KnowledgeBaseListResult struct {
	Total int             `json:"total"`
	List  []KnowledgeBase `json:"list"`
}

func NewCmdKnowledgeList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List knowledge bases",
		Annotations: map[string]string{
			permission.RequiredKey: permission.KnowledgeRead.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if runF != nil {
				return runF(opts)
			}
			return listKBRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")

	return cmd
}

func listKBRun(opts *ListOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(opts.Ctx, "/api/cli/knowledge/list", nil)
	if err != nil {
		return fmt.Errorf("failed to list knowledge bases: %w", err)
	}

	var result KnowledgeBaseListResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	if len(result.List) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No knowledge bases found.")
		return nil
	}

	headers := []string{"CODE", "NAME"}
	rows := make([][]string, 0, len(result.List))
	for _, kb := range result.List {
		rows = append(rows, []string{kb.Code, kb.Name})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)
	fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\nTotal: %d\n", result.Total)

	return nil
}
