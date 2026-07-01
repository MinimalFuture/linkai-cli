package audio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type SpeechOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	DryRun  bool
	Text    string
	Model   string
	Voice   string
	Output  string
}

type AudioSpeechResult struct {
	URL string `json:"url"`
}

func NewCmdAudioSpeech(f *cmdutil.Factory, runF func(*SpeechOptions) error) *cobra.Command {
	opts := &SpeechOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "speech <text>",
		Short: "Generate speech audio from text (TTS)",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.AudioGen.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Text = args[0]
			if runF != nil {
				return runF(opts)
			}
			return speechRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.Model, "model", "tts-1", "TTS model (tts-1 or tts-1-hd)")
	cmd.Flags().StringVar(&opts.Voice, "voice", "", "voice type ID")
	cmd.Flags().StringVar(&opts.Output, "output", "", "save audio to local file (e.g. speech.mp3)")

	return cmd
}

func speechRun(opts *SpeechOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"text":  opts.Text,
		"model": opts.Model,
	}
	if opts.Voice != "" {
		body["voice"] = opts.Voice
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/audio/speech",
			Body:   body,
		})
	}

	fmt.Fprintln(opts.Factory.IOStreams.ErrOut, "Generating speech...")

	resp, err := client.Post(opts.Ctx, "/cli/audio/speech", body)
	if err != nil {
		return fmt.Errorf("failed to generate speech: %w", err)
	}

	var result AudioSpeechResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.Output != "" {
		if err := downloadFile(opts.Ctx, result.URL, opts.Output); err != nil {
			return fmt.Errorf("failed to download audio: %w", err)
		}
		if opts.JSON {
			return output.PrintJSON(opts.Factory.IOStreams.Out, map[string]string{
				"url":  result.URL,
				"file": opts.Output,
			})
		}
		fmt.Fprintf(opts.Factory.IOStreams.Out, "Saved to %s\n", opts.Output)
		return nil
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintln(opts.Factory.IOStreams.Out, result.URL)
	return nil
}

func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
