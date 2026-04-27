package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type RunOptions struct {
	Factory   *cmdutil.Factory
	Ctx       context.Context
	JSON      bool
	AppCode   string
	Input     string
	Args      []string
	SessionID string
}

type WorkflowRunResult struct {
	OutputText string `json:"output_text"`
	SessionID  string `json:"session_id"`
}

func NewCmdWorkflowRun(f *cmdutil.Factory, runF func(*RunOptions) error) *cobra.Command {
	opts := &RunOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "run <app_code>",
		Short: "Run a workflow",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.WorkflowRun.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.AppCode = args[0]
			if runF != nil {
				return runF(opts)
			}
			return runWorkflow(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.Input, "input", "", "input text (input_text)")
	cmd.Flags().StringArrayVar(&opts.Args, "arg", nil, "extra argument in key=value format (can be repeated)")
	cmd.Flags().StringVar(&opts.SessionID, "session", "", "session ID for multi-turn conversation")
	_ = cmd.MarkFlagRequired("input")

	return cmd
}

func runWorkflow(opts *RunOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	extraArgs := map[string]interface{}{}
	for _, a := range opts.Args {
		parts := strings.SplitN(a, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --arg format %q: expected key=value", a)
		}
		extraArgs[parts[0]] = parts[1]
	}

	body := map[string]interface{}{
		"app_code": opts.AppCode,
		"input":    opts.Input,
		"args":     extraArgs,
	}
	if opts.SessionID != "" {
		body["session_id"] = opts.SessionID
	}

	fmt.Fprintln(opts.Factory.IOStreams.ErrOut, "Running workflow...")

	resp, err := client.Post(opts.Ctx, "/api/cli/workflow/run", body)
	if err != nil {
		return fmt.Errorf("failed to run workflow: %w", err)
	}

	var result WorkflowRunResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintln(opts.Factory.IOStreams.Out, result.OutputText)
	if result.SessionID != "" {
		fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\nsession: %s\n", result.SessionID)
	}

	return nil
}
