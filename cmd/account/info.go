package account

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

// InfoOptions holds all inputs for account info.
type InfoOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
}

// AccountInfo represents the response from the account info endpoint.
type AccountInfo struct {
	UserName string `json:"user_name"`
	UserType string `json:"user_type"`
	Score    int64  `json:"score"`
}

// NewCmdAccountInfo creates the account info subcommand.
func NewCmdAccountInfo(f *cmdutil.Factory, runF func(*InfoOptions) error) *cobra.Command {
	opts := &InfoOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show account information",
		Long:  "Display the current user's name, remaining credits, and plan version.",
		Annotations: map[string]string{
			permission.RequiredKey: permission.UserRead.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if runF != nil {
				return runF(opts)
			}
			return infoRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")

	return cmd
}

func infoRun(opts *InfoOptions) error {
	f := opts.Factory

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(opts.Ctx, "/cli/account/info", nil)
	if err != nil {
		return fmt.Errorf("failed to get account info: %w", err)
	}

	var info AccountInfo
	if err := resp.Decode(&info); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(f.IOStreams.Out, info)
	}

	fmt.Fprintf(f.IOStreams.Out, "User:     %s\n", info.UserName)
	fmt.Fprintf(f.IOStreams.Out, "Version:  %s\n", info.UserType)
	fmt.Fprintf(f.IOStreams.Out, "Credits:  %d\n", info.Score)

	return nil
}
