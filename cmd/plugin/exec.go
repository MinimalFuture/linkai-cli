package plugin

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

type ExecOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	Code    string
	Input   string
	Args    []string
}

func NewCmdPluginExec(f *cmdutil.Factory, runF func(*ExecOptions) error) *cobra.Command {
	opts := &ExecOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "exec <code>",
		Short: "Execute a plugin",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "plugin:run",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return execRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.Input, "input", "", "input text for the plugin")
	cmd.Flags().StringArrayVar(&opts.Args, "arg", nil, "structured argument in key=value format (can be repeated)")

	return cmd
}

func execRun(opts *ExecOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"code":  opts.Code,
		"input": opts.Input,
	}

	if len(opts.Args) > 0 {
		argsMap := map[string]interface{}{}
		for _, a := range opts.Args {
			parts := strings.SplitN(a, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid --arg format %q: expected key=value", a)
			}
			argsMap[parts[0]] = parts[1]
		}
		body["args"] = argsMap
	}

	fmt.Fprintln(opts.Factory.IOStreams.ErrOut, "Executing plugin...")

	resp, err := client.Post(opts.Ctx, "/api/cli/plugin/execute", body)
	if err != nil {
		return fmt.Errorf("failed to execute plugin: %w", err)
	}

	var result interface{}
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	switch v := result.(type) {
	case string:
		fmt.Fprintln(opts.Factory.IOStreams.Out, v)
	default:
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	return nil
}
