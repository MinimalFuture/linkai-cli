package knowledge

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type ImportOptions struct {
	Factory   *cmdutil.Factory
	Ctx       context.Context
	JSON      bool
	DryRun    bool
	Code      string
	File      string
	Type      string
	ChunkSize int
}

// validImportTypes are the file kinds the backend accepts, matching the
// platform's import options.
var validImportTypes = map[string]string{
	"doc":   "unstructured document (pdf, txt, word, md, ...)",
	"qa":    "question/answer pairs (csv/excel, two columns)",
	"table": "tabular data (excel/csv, multiple columns)",
}

func NewCmdKnowledgeImport(f *cmdutil.Factory, runF func(*ImportOptions) error) *cobra.Command {
	opts := &ImportOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "import <kb_code> --file <path> [--type doc|qa|table]",
		Short: "Import a local file into a knowledge base",
		Long: "Import a local file into a knowledge base, same as the platform's file import. " +
			"Types:\n" +
			"  doc   — unstructured document (pdf, txt, word, md, ...)  [default]\n" +
			"  qa    — question/answer pairs (csv or excel, two columns)\n" +
			"  table — tabular data (excel or csv, multiple columns)\n\n" +
			"The file is uploaded and embedded asynchronously; the command returns once the " +
			"import has been accepted. Use `linkai knowledge files <kb_code>` to see it appear.",
		Example: `  linkai knowledge import kb_abc --file ./manual.pdf
  linkai knowledge import kb_abc --file ./faq.xlsx --type qa
  linkai knowledge import kb_abc --file ./data.csv --type table --json`,
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
			return importRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.File, "file", "", "path to the local file to import (required)")
	cmd.Flags().StringVar(&opts.Type, "type", "doc", "file kind: doc | qa | table")
	cmd.Flags().IntVar(&opts.ChunkSize, "chunk-size", 1000, "chunk size for splitting (doc only)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func importRun(opts *ImportOptions) error {
	if _, ok := validImportTypes[opts.Type]; !ok {
		return output.ErrValidation("invalid --type %q: expected doc, qa, or table", opts.Type)
	}
	if _, err := os.Stat(opts.File); err != nil {
		return output.ErrValidation("cannot access file %q: %v", opts.File, err)
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/knowledge/file/import",
			Body: map[string]interface{}{
				"code":       opts.Code,
				"type":       opts.Type,
				"chunk_size": opts.ChunkSize,
				"file":       opts.File,
			},
		})
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	fields := map[string]string{
		"code":       opts.Code,
		"type":       opts.Type,
		"chunk_size": strconv.Itoa(opts.ChunkSize),
	}
	resp, err := client.PostMultipart(opts.Ctx, "/cli/knowledge/file/import", fields, "file", opts.File)
	if err != nil {
		return fmt.Errorf("failed to import file: %w", err)
	}

	var result map[string]interface{}
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintf(opts.Factory.IOStreams.Out, "Import accepted: %s (type: %s) → knowledge base %s\n", opts.File, opts.Type, opts.Code)
	fmt.Fprintln(opts.Factory.IOStreams.ErrOut, "Embedding runs in the background; run `linkai knowledge files "+opts.Code+"` to check progress.")
	return nil
}
