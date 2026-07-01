package knowledge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
	"github.com/MinimalFuture/linkai-cli/internal/validate"
)

type AddOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	DryRun  bool

	Code     string
	FileID   string
	Text     string
	Question string
	Answer   string
}

func NewCmdKnowledgeAdd(f *cmdutil.Factory, runF func(*AddOptions) error) *cobra.Command {
	opts := &AddOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "add <code>",
		Short: "Add a text chunk or QA entry to a knowledge base",
		Long: "Add content to a knowledge base. Provide --text for a raw chunk, " +
			"or --question/--answer for a QA entry. Use --file-id to append to an " +
			"existing file; when omitted a new file is created and its id returned.",
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.KnowledgeCreate.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return addRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.FileID, "file-id", "", "append to this file (default: create a new file)")
	cmd.Flags().StringVar(&opts.Text, "text", "", "raw text chunk")
	cmd.Flags().StringVar(&opts.Question, "question", "", "QA question (requires --answer)")
	cmd.Flags().StringVar(&opts.Answer, "answer", "", "QA answer (requires --question)")

	return cmd
}

func addRun(opts *AddOptions) error {
	isQA := opts.Question != ""
	if isQA {
		if opts.Answer == "" {
			return output.ErrValidation("--answer is required when using --question")
		}
	} else if opts.Text == "" {
		return output.ErrValidation("provide --text, or --question with --answer")
	}

	for field, val := range map[string]string{"text": opts.Text, "question": opts.Question, "answer": opts.Answer} {
		if val != "" {
			if err := validate.RejectControlChars(field, val); err != nil {
				return err
			}
		}
	}

	body := map[string]interface{}{"code": opts.Code}
	if opts.FileID != "" {
		body["file_id"] = opts.FileID
	}
	if isQA {
		body["question"] = opts.Question
		body["answer"] = opts.Answer
	} else {
		body["text"] = opts.Text
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/knowledge/data/add",
			Body:   body,
		})
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Post(opts.Ctx, "/cli/knowledge/data/add", body)
	if err != nil {
		return fmt.Errorf("failed to add knowledge data: %w", err)
	}

	var result struct {
		FileID string `json:"file_id"`
	}
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintf(opts.Factory.IOStreams.Out, "Content added to knowledge base %q (file: %s)\n", opts.Code, result.FileID)
	return nil
}
