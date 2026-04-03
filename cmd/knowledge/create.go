package knowledge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yjr/linkai-cli/internal/cmdutil"
	"github.com/yjr/linkai-cli/internal/output"
)

type CreateOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	Name    string
	Desc    string
}

func NewCmdKnowledgeCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a knowledge base",
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "knowledge:write",
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
