package image

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

type GenOptions struct {
	Factory     *cmdutil.Factory
	Ctx         context.Context
	JSON        bool
	DryRun      bool
	Prompt      string
	Model       string
	Size        string
	AspectRatio string
}

type ImageGenResult struct {
	URL string `json:"url"`
}

func NewCmdImageGen(f *cmdutil.Factory, runF func(*GenOptions) error) *cobra.Command {
	opts := &GenOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "gen <prompt>",
		Short: "Generate an image from a text prompt",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "image:write",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Prompt = args[0]
			if runF != nil {
				return runF(opts)
			}
			return genRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.Model, "model", "", "image model (e.g. dall-e-3, doubao-seedream-4.5)")
	cmd.Flags().StringVar(&opts.Size, "size", "", "image size (e.g. 1024x1024)")
	cmd.Flags().StringVar(&opts.AspectRatio, "aspect-ratio", "", "aspect ratio (e.g. 1:1, 16:9)")

	return cmd
}

func genRun(opts *GenOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"prompt": opts.Prompt,
	}
	if opts.Model != "" {
		body["model"] = opts.Model
	}
	if opts.Size != "" {
		body["size"] = opts.Size
	}
	if opts.AspectRatio != "" {
		body["aspect_ratio"] = opts.AspectRatio
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/api/cli/image/gen",
			Body:   body,
		})
	}

	fmt.Fprintln(opts.Factory.IOStreams.ErrOut, "Generating image...")

	resp, err := client.Post(opts.Ctx, "/api/cli/image/gen", body)
	if err != nil {
		return fmt.Errorf("failed to generate image: %w", err)
	}

	var result ImageGenResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintln(opts.Factory.IOStreams.Out, result.URL)
	return nil
}
