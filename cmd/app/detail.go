package app

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type DetailOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	Code    string
}

type AppDetail struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func NewCmdAppDetail(f *cmdutil.Factory, runF func(*DetailOptions) error) *cobra.Command {
	opts := &DetailOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "detail <code>",
		Short: "View application detail",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.AppRead.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return detailRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")

	return cmd
}

func detailRun(opts *DetailOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("code", opts.Code)

	resp, err := client.Get(opts.Ctx, "/api/cli/app/detail", params)
	if err != nil {
		return fmt.Errorf("failed to get app detail: %w", err)
	}

	var detail AppDetail
	if err := resp.Decode(&detail); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, detail)
	}

	w := opts.Factory.IOStreams.Out
	fmt.Fprintf(w, "Code:        %s\n", detail.Code)
	fmt.Fprintf(w, "Name:        %s\n", detail.Name)
	fmt.Fprintf(w, "Type:        %s\n", detail.Type)
	if detail.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", detail.Description)
	}

	return nil
}
