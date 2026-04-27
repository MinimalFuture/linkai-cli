package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type ListOptions struct {
	Factory       *cmdutil.Factory
	Ctx           context.Context
	JSON          bool
	ModelType     string
	SupplierType  string
}

type ModelItem struct {
	ModelCode  string   `json:"modelCode"`
	ModelTypes []string `json:"modelTypes"`
	ModelName  string   `json:"modelName"`
}

func NewCmdModelList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available AI models",
		Annotations: map[string]string{
			permission.RequiredKey: permission.AppRead.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.ModelType, "type", "", "filter by model type (e.g. LLM, EMBEDDING)")
	cmd.Flags().StringVar(&opts.SupplierType, "supplier", "", "filter by supplier code (e.g. openai, claude)")

	return cmd
}

func listRun(opts *ListOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{}
	if opts.ModelType != "" {
		body["modelTypeList"] = []string{opts.ModelType}
	}
	if opts.SupplierType != "" {
		body["modelSupplierTypeList"] = []string{opts.SupplierType}
	}

	resp, err := client.Post(opts.Ctx, "/api/cli/model/list", body)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	var models []ModelItem
	if err := resp.Decode(&models); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, models)
	}

	if len(models) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No models found.")
		return nil
	}

	headers := []string{"CODE", "NAME", "TYPES"}
	rows := make([][]string, 0, len(models))
	for _, m := range models {
		types := "-"
		if len(m.ModelTypes) > 0 {
			types = strings.Join(m.ModelTypes, " / ")
		}
		rows = append(rows, []string{m.ModelCode, m.ModelName, types})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)
	fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\nTotal: %d\n", len(models))

	return nil
}
