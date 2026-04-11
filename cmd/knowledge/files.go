package knowledge

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

type FilesOptions struct {
	Factory  *cmdutil.Factory
	Ctx      context.Context
	JSON     bool
	Code     string
	FileName string
	Page     int
	PageSize int
}

type KnowledgeFile struct {
	FileID   string `json:"fileId"`
	FileName string `json:"fileName"`
	DataType string `json:"dataType"`
	Status   string `json:"status"`
	Total    int    `json:"total"`
}

type KnowledgeFileListResult struct {
	Total    int             `json:"total"`
	PageNum  int             `json:"pageNum"`
	PageSize int             `json:"pageSize"`
	List     []KnowledgeFile `json:"list"`
}

func NewCmdKnowledgeFiles(f *cmdutil.Factory, runF func(*FilesOptions) error) *cobra.Command {
	opts := &FilesOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "files <code>",
		Short: "List files in a knowledge base",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "knowledge:read",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return filesRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.FileName, "name", "", "filter by file name")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "page number")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "number of items per page")

	return cmd
}

func filesRun(opts *FilesOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("code", opts.Code)
	params.Set("pageNo", strconv.Itoa(opts.Page))
	params.Set("pageSize", strconv.Itoa(opts.PageSize))
	if opts.FileName != "" {
		params.Set("file_name", opts.FileName)
	}

	resp, err := client.Get(opts.Ctx, "/api/cli/knowledge/files", params)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	var result KnowledgeFileListResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	if len(result.List) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No files found.")
		return nil
	}

	headers := []string{"FILE ID", "FILE NAME", "TYPE", "STATUS", "CHUNKS"}
	rows := make([][]string, 0, len(result.List))
	for _, f := range result.List {
		name := string([]rune(f.FileName))
		runes := []rune(name)
		if len(runes) > 40 {
			name = string(runes[:40]) + "..."
		}
		rows = append(rows, []string{f.FileID, name, f.DataType, f.Status, strconv.Itoa(f.Total)})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)

	start := (result.PageNum-1)*result.PageSize + 1
	end := start + len(result.List) - 1
	fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\nShowing %d-%d of %d\n", start, end, result.Total)

	return nil
}
