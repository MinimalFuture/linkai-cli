package database

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type DescribeOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	Code    string
	Table   string
}

type FieldItem struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment"`
}

type DescribeResult struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Fields      []FieldItem `json:"fields"`
}

func NewCmdDatabaseDescribe(f *cmdutil.Factory, runF func(*DescribeOptions) error) *cobra.Command {
	opts := &DescribeOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "describe <code> <table>",
		Short: "Show table structure",
		Args:  cobra.ExactArgs(2),
		Annotations: map[string]string{
			permission.RequiredKey: permission.DBRead.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			opts.Table = args[1]
			if runF != nil {
				return runF(opts)
			}
			return describeRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")

	return cmd
}

func describeRun(opts *DescribeOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("code", opts.Code)
	params.Set("table", opts.Table)

	resp, err := client.Get(opts.Ctx, "/cli/database/describe", params)
	if err != nil {
		return fmt.Errorf("failed to describe table: %w", err)
	}

	var result DescribeResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	if result.Description != "" {
		fmt.Fprintf(opts.Factory.IOStreams.Out, "Table: %s  (%s)\n\n", result.Name, result.Description)
	} else {
		fmt.Fprintf(opts.Factory.IOStreams.Out, "Table: %s\n\n", result.Name)
	}

	if len(result.Fields) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No fields found.")
		return nil
	}

	headers := []string{"COLUMN", "TYPE", "COMMENT"}
	rows := make([][]string, 0, len(result.Fields))
	for _, f := range result.Fields {
		rows = append(rows, []string{f.Name, f.Type, f.Comment})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)

	return nil
}
