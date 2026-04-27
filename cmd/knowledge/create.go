package knowledge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type CreateOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	DryRun  bool
	Name    string
	Desc    string
}

func NewCmdKnowledgeCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a knowledge base",
		Annotations: map[string]string{
			permission.RequiredKey: permission.KnowledgeCreate.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.Name, "name", "", "knowledge base name (required)")
	cmd.Flags().StringVar(&opts.Desc, "desc", "", "knowledge base description")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func createRun(opts *CreateOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"name": opts.Name,
		"desc": opts.Desc,
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/api/cli/knowledge/create",
			Body:   body,
		})
	}

	resp, err := client.Post(opts.Ctx, "/api/cli/knowledge/create", body)
	if err != nil {
		return fmt.Errorf("failed to create knowledge base: %w", err)
	}

	var result map[string]interface{}
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	code, _ := result["code"].(string)
	fmt.Fprintf(opts.Factory.IOStreams.Out, "Knowledge base created: %s (code: %s)\n", opts.Name, code)
	return nil
}
