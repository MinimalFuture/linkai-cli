package plugin

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

type DetailOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	Code    string
}

type PluginDetail struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	ShortDesc string `json:"shortDesc"`
	Desc      string `json:"desc"`
	Intro     string `json:"intro"`
	ExtParam  string `json:"extParam"`
}

func NewCmdPluginDetail(f *cmdutil.Factory, runF func(*DetailOptions) error) *cobra.Command {
	opts := &DetailOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "detail <code>",
		Short: "Show plugin detail",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "plugin:read",
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

	resp, err := client.Get(opts.Ctx, "/api/cli/plugin/detail", params)
	if err != nil {
		return fmt.Errorf("failed to get plugin detail: %w", err)
	}

	var detail PluginDetail
	if err := resp.Decode(&detail); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, detail)
	}

	w := opts.Factory.IOStreams.Out
	fmt.Fprintf(w, "Name:     %s\n", detail.Name)
	fmt.Fprintf(w, "Code:     %s\n", detail.Code)
	if detail.ShortDesc != "" {
		fmt.Fprintf(w, "Summary:  %s\n", detail.ShortDesc)
	}
	if detail.Desc != "" {
		fmt.Fprintf(w, "Desc:     %s\n", detail.Desc)
	}
	if detail.Intro != "" {
		fmt.Fprintf(w, "\n%s\n", detail.Intro)
	}
	if detail.ExtParam != "" {
		fmt.Fprintf(w, "\nExtParam: %s\n", detail.ExtParam)
	}

	return nil
}
