package plugin

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

type ListOptions struct {
	Factory  *cmdutil.Factory
	Ctx      context.Context
	JSON     bool
	Category string
}

type PluginItem struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	ShortDesc string `json:"shortDesc"`
	Category  string `json:"category"`
	ChargeType string `json:"chargeType"`
}

func NewCmdPluginList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available plugins",
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "plugin:read",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.Category, "category", "", "filter by category")

	return cmd
}

func listRun(opts *ListOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	if opts.Category != "" {
		params.Set("category", opts.Category)
	}

	resp, err := client.Get(opts.Ctx, "/api/cli/plugin/list", params)
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	var items []PluginItem
	if err := resp.Decode(&items); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, items)
	}

	if len(items) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No plugins found.")
		return nil
	}

	headers := []string{"CODE", "NAME", "CATEGORY", "DESCRIPTION"}
	rows := make([][]string, 0, len(items))
	for _, p := range items {
		desc := []rune(p.ShortDesc)
		if len(desc) > 40 {
			desc = append(desc[:40], []rune("...")...)
		}
		rows = append(rows, []string{p.Code, p.Name, p.Category, string(desc)})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)

	return nil
}
