package video

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type VideoGenOptions struct {
	Factory     *cmdutil.Factory
	Ctx         context.Context
	JSON        bool
	DryRun      bool
	Prompt      string
	Model       string
	Duration    int
	AspectRatio string
	Size        string
	Mode        string
	Images      []string
	ImageMode   string
}

type VideoTaskResult struct {
	TaskID       string `json:"task_id"`
	Status       string `json:"status"`
	Model        string `json:"model"`
	VideoURL     string `json:"video_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func NewCmdVideoGen(f *cmdutil.Factory, runF func(*VideoGenOptions) error) *cobra.Command {
	opts := &VideoGenOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "gen <prompt>",
		Short: "Generate a video from a text prompt",
		Long: `Generate a video from a text prompt (text-to-video), or from a
reference image by passing --image (image-to-video).

The CLI submits the task and polls for completion, printing the video URL
when ready. Generation typically takes 30s–3 minutes depending on the model.

Models are not hardcoded: when --model is omitted the platform picks the
account's default video model. Discover the available video models with:

  linkai model list --type VIDEO

All sizing options (duration, aspect ratio, size, mode) are model-specific and
validated by the server; values left unset fall back to the model's default.`,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.VideoGen.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Prompt = args[0]
			if runF != nil {
				return runF(opts)
			}
			return videoGenRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.Model, "model", "", "video model code; omit to use the account default (see: linkai model list --type VIDEO)")
	cmd.Flags().IntVar(&opts.Duration, "duration", 0, "video duration in seconds; 0 = model default (range is model-specific)")
	cmd.Flags().StringVar(&opts.AspectRatio, "aspect-ratio", "", "aspect ratio; values depend on the model (e.g. 16:9, 9:16, 1:1)")
	cmd.Flags().StringVar(&opts.Size, "size", "", "video resolution; values depend on the model (e.g. 480P, 720P, 1080P)")
	cmd.Flags().StringVar(&opts.Mode, "mode", "", "generation mode (kling models only): std or pro")
	cmd.Flags().StringArrayVar(&opts.Images, "image", nil, "reference image URL for image-to-video; repeat for multiple (model-dependent)")
	cmd.Flags().StringVar(&opts.ImageMode, "image-mode", "", "how to use --image: reference or first_last_frame (model-dependent)")

	return cmd
}

func videoGenRun(opts *VideoGenOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	// Step 1: create task. Only forward flags the user explicitly set so the
	// server/model decides every default the CLI doesn't know about.
	createBody := map[string]interface{}{
		"prompt": opts.Prompt,
	}
	if opts.Model != "" {
		createBody["model"] = opts.Model
	}
	if opts.Duration > 0 {
		createBody["duration"] = opts.Duration
	}
	if opts.AspectRatio != "" {
		createBody["aspect_ratio"] = opts.AspectRatio
	}
	if opts.Size != "" {
		createBody["size"] = opts.Size
	}
	if opts.Mode != "" {
		createBody["mode"] = opts.Mode
	}
	if len(opts.Images) > 0 {
		createBody["images"] = opts.Images
	}
	if opts.ImageMode != "" {
		createBody["image_mode"] = opts.ImageMode
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/api/cli/video/gen",
			Body:   createBody,
		})
	}

	resp, err := client.Post(opts.Ctx, "/api/cli/video/gen", createBody)
	if err != nil {
		return fmt.Errorf("failed to create video task: %w", err)
	}

	var task VideoTaskResult
	if err := resp.Decode(&task); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "Video task created (id: %s), waiting for completion...\n", task.TaskID)

	// Step 2: poll for status. The server resolves model/duration/mode from the
	// task metadata, so the query only needs the task_id.
	const pollInterval = 5 * time.Second
	const timeout = 10 * time.Minute
	start := time.Now()
	deadline := start.Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-time.After(pollInterval):
		case <-opts.Ctx.Done():
			return fmt.Errorf("canceled while waiting for video (task_id: %s)", task.TaskID)
		}

		statusResp, err := client.Post(opts.Ctx, "/api/cli/video/status", map[string]interface{}{
			"task_id": task.TaskID,
		})
		if err != nil {
			return fmt.Errorf("failed to query video status: %w", err)
		}

		var status VideoTaskResult
		if err := statusResp.Decode(&status); err != nil {
			return fmt.Errorf("failed to parse status response: %w", err)
		}

		elapsed := time.Since(start).Round(time.Second)
		fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "Status: %s  [%s elapsed]\n", status.Status, elapsed)

		switch status.Status {
		case "completed", "success":
			if opts.JSON {
				return output.PrintJSON(opts.Factory.IOStreams.Out, status)
			}
			fmt.Fprintln(opts.Factory.IOStreams.Out, status.VideoURL)
			return nil
		case "failed", "error":
			if status.ErrorMessage != "" {
				return fmt.Errorf("video generation failed (task_id: %s): %s", task.TaskID, status.ErrorMessage)
			}
			return fmt.Errorf("video generation failed (task_id: %s)", task.TaskID)
		}
		// init / queued / processing / pending — keep polling
	}

	return fmt.Errorf("timed out waiting for video (task_id: %s) — generation may still be in progress", task.TaskID)
}
