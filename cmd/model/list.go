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
		Long: `List the AI models available to your account.

Use --type to discover models for a specific capability, e.g.:

  linkai model list --type IMAGE   # image generation models (for: linkai image gen --model)
  linkai model list --type VIDEO   # video generation models (for: linkai video gen --model)
  linkai model list --type LLM     # large language models

The model codes printed here are the values to pass to the --model flag of the
corresponding command. Available models change over time, so prefer querying
this list rather than hardcoding a model name.`,
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
	cmd.Flags().StringVar(&opts.ModelType, "type", "", "filter by model type (e.g. IMAGE, VIDEO, LLM, EMBEDDING)")
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
		if canonical := canonicalTypes(m.ModelTypes); len(canonical) > 0 {
			types = strings.Join(canonical, " / ")
		}
		rows = append(rows, []string{m.ModelCode, m.ModelName, types})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)
	fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\nTotal: %d\n", len(models))

	return nil
}

// canonicalTypes normalizes the backend's model-type labels (a mix of Chinese
// and English strings) into a small, stable, English-only set that is friendly
// to both overseas users and agents:
//
//	LLM       — language models (chat / reasoning are merged: reasoning is just
//	            an LLM capability, not a separate category)
//	IMAGE     — image generation / understanding
//	VIDEO     — video generation
//	EMBEDDING — embedding models
//
// Unknown labels are passed through unchanged so we never silently hide a new
// backend category. The result is de-duplicated while preserving first-seen
// order.
func canonicalTypes(raw []string) []string {
	seen := make(map[string]bool, len(raw))
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		c := canonicalType(t)
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	return out
}

func canonicalType(t string) string {
	switch strings.TrimSpace(t) {
	case "大语言模型", "推理模型", "REASONER", "LLM":
		return "LLM"
	case "图像生成模型", "图像理解模型", "IMAGE":
		return "IMAGE"
	case "VIDEO", "视频生成模型":
		return "VIDEO"
	case "EMBEDDING", "向量模型", "嵌入模型":
		return "EMBEDDING"
	default:
		return strings.TrimSpace(t)
	}
}
