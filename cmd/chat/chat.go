package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/api"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type ChatOptions struct {
	Factory   *cmdutil.Factory
	Ctx       context.Context
	JSON      bool
	DryRun    bool
	NoStream  bool
	App       string
	SessionID string
	Message   string
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	SessionID string `json:"session_id"`
}

type chatResult struct {
	Reply     string `json:"reply"`
	SessionID string `json:"session_id"`
}

func NewCmdChat(f *cmdutil.Factory, runF func(*ChatOptions) error) *cobra.Command {
	opts := &ChatOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "chat <message>",
		Short: "Chat with an application",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.ChatSend.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Message = args[0]
			// Streaming default is context-aware: a human at a terminal gets the
			// typewriter effect, but when the output is piped/redirected (the
			// typical AI-agent-via-bash case) we default to non-streaming so the
			// reply arrives as one clean block. --stream / --no-stream override
			// this, and --json always forces non-streaming.
			if opts.JSON {
				opts.NoStream = true
			} else if !cmd.Flags().Changed("stream") && !cmd.Flags().Changed("no-stream") {
				opts.NoStream = !f.IOStreams.IsTerminal
			}
			if runF != nil {
				return runF(opts)
			}
			return chatRun(opts)
		},
	}

	var stream bool
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format (disables streaming)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().BoolVar(&opts.NoStream, "no-stream", false, "disable streaming, wait for full reply")
	cmd.Flags().BoolVar(&stream, "stream", false, "force streaming output (default when stdout is a terminal)")
	cmd.MarkFlagsMutuallyExclusive("stream", "no-stream")
	cmd.Flags().StringVar(&opts.App, "app", "", "application code")
	cmd.Flags().StringVar(&opts.SessionID, "session", "", "session ID for multi-turn conversation")
	_ = cmd.MarkFlagRequired("app")

	return cmd
}

func chatRun(opts *ChatOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"app_code": opts.App,
		"message":  opts.Message,
		"stream":   !opts.NoStream,
	}
	if opts.SessionID != "" {
		body["session_id"] = opts.SessionID
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/api/cli/chat/completions",
			Body:   body,
		})
	}

	if opts.NoStream {
		return chatNoStream(opts, client, body)
	}
	return chatStream(opts, client, body)
}

func chatStream(opts *ChatOptions, client *api.Client, body map[string]interface{}) error {
	stream, err := client.Stream(opts.Ctx, "/api/cli/chat/completions", body)
	if err != nil {
		return fmt.Errorf("failed to start chat: %w", err)
	}
	defer stream.Close()

	out := opts.Factory.IOStreams.Out
	errOut := opts.Factory.IOStreams.ErrOut
	var sessionID string

	const maxSSELineSize = 1024 * 1024 // 1 MB
	scanner := bufio.NewScanner(stream)
	scanner.Buffer(make([]byte, 0, 64*1024), maxSSELineSize)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line == "data: [DONE]" {
			continue
		}
		data, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			snippet := data
			if len(snippet) > 120 {
				snippet = snippet[:120] + "…"
			}
			fmt.Fprintf(errOut, "[linkai] [WARN] failed to parse SSE chunk: %v (data: %s)\n", err, snippet)
			continue
		}

		if len(chunk.Choices) > 0 {
			fmt.Fprint(out, chunk.Choices[0].Delta.Content)
		}
		if chunk.SessionID != "" {
			sessionID = chunk.SessionID
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream read error: %w", err)
	}

	fmt.Fprintln(out)
	if sessionID != "" {
		fmt.Fprintf(errOut, "\nsession: %s\n", sessionID)
	}
	return nil
}

func chatNoStream(opts *ChatOptions, client *api.Client, body map[string]interface{}) error {
	resp, err := client.Post(opts.Ctx, "/api/cli/chat/completions", body)
	if err != nil {
		return fmt.Errorf("failed to chat: %w", err)
	}

	var result chatResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintln(opts.Factory.IOStreams.Out, result.Reply)
	if result.SessionID != "" {
		fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\nsession: %s\n", result.SessionID)
	}
	return nil
}
