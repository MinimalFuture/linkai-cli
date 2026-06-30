package image

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
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
	Quality     string
	Images      []string
}

type ImageGenResult struct {
	URL string `json:"url"`
}

func NewCmdImageGen(f *cmdutil.Factory, runF func(*GenOptions) error) *cobra.Command {
	opts := &GenOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "gen <prompt>",
		Short: "Generate an image from a text prompt",
		Long: `Generate an image from a text prompt (text-to-image), or edit
reference images by passing --image (image-to-image).

Models are not hardcoded: when --model is omitted the platform picks the
account's default image model. Discover the available image models with:

  linkai model list --type IMAGE

Available sizes / aspect ratios / quality depend on the chosen model and are
validated by the server; unsupported fields are ignored rather than rejected.`,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.ImageGen.String(),
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
	cmd.Flags().StringVar(&opts.Model, "model", "", "image model code; omit to use the account default (see: linkai model list --type IMAGE)")
	cmd.Flags().StringVar(&opts.Size, "size", "", "image size; values depend on the model (e.g. 1K, 2K, 4K)")
	cmd.Flags().StringVar(&opts.AspectRatio, "aspect-ratio", "", "aspect ratio; values depend on the model (e.g. 1:1, 16:9, 9:16)")
	cmd.Flags().StringVar(&opts.Quality, "quality", "", "image quality; values depend on the model (e.g. standard, hd)")
	cmd.Flags().StringArrayVar(&opts.Images, "image", nil, "reference image URL for image-to-image; repeat for multiple (model-dependent)")

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
	// Only forward flags the user explicitly set, so the server/model decides
	// every default the CLI doesn't know about.
	if opts.Model != "" {
		body["model"] = opts.Model
	}
	if opts.Size != "" {
		body["size"] = opts.Size
	}
	if opts.AspectRatio != "" {
		body["aspect_ratio"] = opts.AspectRatio
	}
	if opts.Quality != "" {
		body["quality"] = opts.Quality
	}
	if len(opts.Images) > 0 {
		body["images"] = opts.Images
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
