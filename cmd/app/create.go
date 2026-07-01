package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
	"github.com/MinimalFuture/linkai-cli/internal/validate"
)

type CreateOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	DryRun  bool

	Name         string
	Type         string
	Desc         string
	Prompt       string
	Introduction string
}

func NewCmdAppCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an application",
		Long: "Create an application. Only --name is required; the type defaults " +
			"to PROMPT (lightweight app) and all other fields use server defaults.",
		Annotations: map[string]string{
			permission.RequiredKey: permission.AppCreate.String(),
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
	cmd.Flags().StringVar(&opts.Name, "name", "", "application name (required)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "application type: PROMPT|EMBEDDING|WORKFLOW|AGENT (default PROMPT)")
	cmd.Flags().StringVar(&opts.Desc, "desc", "", "application description")
	cmd.Flags().StringVar(&opts.Prompt, "prompt", "", "system prompt / persona")
	cmd.Flags().StringVar(&opts.Introduction, "intro", "", "opening message")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func createRun(opts *CreateOptions) error {
	if err := validate.RejectControlChars("name", opts.Name); err != nil {
		return err
	}

	body := map[string]interface{}{"name": opts.Name}
	if opts.Type != "" {
		body["type"] = opts.Type
	}
	if opts.Desc != "" {
		body["description"] = opts.Desc
	}
	if opts.Prompt != "" {
		body["prompt"] = opts.Prompt
	}
	if opts.Introduction != "" {
		body["introduction"] = opts.Introduction
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/app/create",
			Body:   body,
		})
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Post(opts.Ctx, "/cli/app/create", body)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	var result struct {
		Code  string                 `json:"code"`
		Name  string                 `json:"name"`
		Links map[string]interface{} `json:"links,omitempty"`
	}
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintf(opts.Factory.IOStreams.Out, "Application created: %s (code: %s)\n", result.Name, result.Code)
	output.PrintLinks(opts.Factory.IOStreams.Out, mapAny(result.Links))
	return nil
}

// mapAny adapts a typed map for PrintLinks which expects interface{}.
func mapAny(m map[string]interface{}) interface{} {
	if m == nil {
		return nil
	}
	return m
}
