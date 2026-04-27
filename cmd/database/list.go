package database

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type ListOptions struct {
	Factory  *cmdutil.Factory
	Ctx      context.Context
	JSON     bool
	Page     int
	PageSize int
}

type DatabaseItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type DatabaseListResult struct {
	Total int            `json:"total"`
	List  []DatabaseItem `json:"list"`
}

func NewCmdDatabaseList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List databases",
		Annotations: map[string]string{
			permission.RequiredKey: permission.DBRead.String(),
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
	cmd.Flags().IntVar(&opts.Page, "page", 1, "page number")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "number of items per page")

	return cmd
}

func listRun(opts *ListOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("pageNo", strconv.Itoa(opts.Page))
	params.Set("pageSize", strconv.Itoa(opts.PageSize))

	resp, err := client.Get(opts.Ctx, "/api/cli/database/list", params)
	if err != nil {
		return fmt.Errorf("failed to list databases: %w", err)
	}

	var result DatabaseListResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	if len(result.List) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No databases found.")
		return nil
	}

	headers := []string{"CODE", "NAME"}
	rows := make([][]string, 0, len(result.List))
	for _, db := range result.List {
		rows = append(rows, []string{db.Code, db.Name})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)

	start := (opts.Page-1)*opts.PageSize + 1
	end := start + len(result.List) - 1
	fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\nShowing %d-%d of %d\n", start, end, result.Total)

	return nil
}
