package app

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

// ListOptions holds all inputs for app list.
type ListOptions struct {
	Factory  *cmdutil.Factory
	Ctx      context.Context
	JSON     bool
	Key      string
	Page     int
	PageSize int
}

// AppItem represents a single application in the list response.
type AppItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// PageResult represents the PageInfo response from the backend.
type PageResult struct {
	Total    int       `json:"total"`
	PageNum  int       `json:"pageNum"`
	PageSize int       `json:"pageSize"`
	List     []AppItem `json:"list"`
}

// NewCmdAppList creates the app list subcommand.
func NewCmdAppList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List applications",
		Long:  "List applications visible to the current user, with optional keyword search and pagination.",
		Annotations: map[string]string{
			permission.RequiredKey: permission.AppRead.String(),
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
	cmd.Flags().StringVar(&opts.Key, "key", "", "search keyword")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "page number")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "number of items per page")

	return cmd
}

func listRun(opts *ListOptions) error {
	f := opts.Factory

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("pageNo", strconv.Itoa(opts.Page))
	params.Set("pageSize", strconv.Itoa(opts.PageSize))
	if opts.Key != "" {
		params.Set("key", opts.Key)
	}

	resp, err := client.Get(opts.Ctx, "/cli/app/list", params)
	if err != nil {
		return fmt.Errorf("failed to list apps: %w", err)
	}

	var page PageResult
	if err := resp.Decode(&page); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(f.IOStreams.Out, page)
	}

	if len(page.List) == 0 {
		fmt.Fprintln(f.IOStreams.Out, "No applications found.")
		return nil
	}

	headers := []string{"CODE", "NAME"}
	rows := make([][]string, 0, len(page.List))
	for _, app := range page.List {
		rows = append(rows, []string{app.Code, app.Name})
	}
	output.PrintTable(f.IOStreams.Out, headers, rows)

	start := (page.PageNum-1)*page.PageSize + 1
	end := start + len(page.List) - 1
	fmt.Fprintf(f.IOStreams.ErrOut, "\nShowing %d-%d of %d\n", start, end, page.Total)

	return nil
}
