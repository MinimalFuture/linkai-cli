package video

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
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
	Mode        string
}

type VideoTaskResult struct {
	TaskID       string `json:"task_id"`
	Status       string `json:"status"`
	Model        string `json:"model"`
	VideoURL     string `json:"video_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

func NewCmdVideoGen(f *cmdutil.Factory, runF func(*VideoGenOptions) error) *cobra.Command {
	opts := &VideoGenOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "gen <prompt>",
		Short: "Generate a video from a text prompt",
		Long: `Generate a video from a text prompt.

The CLI automatically polls for completion and prints the video URL when ready.
Generation typically takes 30s–3 minutes depending on the model.`,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "video:write",
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
	cmd.Flags().StringVar(&opts.Model, "model", "", "video model (e.g. jimeng_t2v_v30)")
	cmd.Flags().IntVar(&opts.Duration, "duration", 5, "video duration in seconds")
	cmd.Flags().StringVar(&opts.AspectRatio, "aspect-ratio", "16:9", "aspect ratio (e.g. 16:9, 9:16, 1:1)")
	cmd.Flags().StringVar(&opts.Mode, "mode", "std", "generation mode: std or pro")

	return cmd
}

func videoGenRun(opts *VideoGenOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	// Step 1: create task
	createBody := map[string]interface{}{
		"prompt":       opts.Prompt,
		"duration":     opts.Duration,
		"aspect_ratio": opts.AspectRatio,
		"mode":         opts.Mode,
	}
	if opts.Model != "" {
		createBody["model"] = opts.Model
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

	// Step 2: poll for status
	const pollInterval = 3 * time.Second
	const timeout = 5 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-time.After(pollInterval):
		case <-opts.Ctx.Done():
			return fmt.Errorf("cancelled while waiting for video (task_id: %s)", task.TaskID)
		}

		statusBody := map[string]interface{}{
			"task_id":  task.TaskID,
			"duration": opts.Duration,
			"mode":     opts.Mode,
		}
		if task.Model != "" {
			statusBody["model"] = task.Model
		}

		statusResp, err := client.Post(opts.Ctx, "/api/cli/video/status", statusBody)
		if err != nil {
			return fmt.Errorf("failed to query video status: %w", err)
		}

		var status VideoTaskResult
		if err := statusResp.Decode(&status); err != nil {
			return fmt.Errorf("failed to parse status response: %w", err)
		}

		elapsed := time.Since(deadline.Add(-timeout)).Round(time.Second)
		fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "Status: %s  [%s elapsed]\n", status.Status, elapsed)

		switch status.Status {
		case "success":
			if opts.JSON {
				return output.PrintJSON(opts.Factory.IOStreams.Out, status)
			}
			fmt.Fprintln(opts.Factory.IOStreams.Out, status.VideoURL)
			return nil
		case "error", "failed":
			return fmt.Errorf("video generation failed (task_id: %s)", task.TaskID)
		}
		// pending / processing — keep polling
	}

	return fmt.Errorf("timed out waiting for video (task_id: %s) — generation may still be in progress", task.TaskID)
}
