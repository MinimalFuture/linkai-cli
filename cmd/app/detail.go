package app

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

type DetailOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	Code    string
}

type AppDetail struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	HeadImageURL string `json:"headImageUrl"`
	Introduction string `json:"introduction"`
	NickName    string `json:"nickName"`
	UsageCount  int    `json:"usageCount"`
	ThumbCount  int    `json:"thumbCount"`
	AppStatus   string `json:"appStatus"`

	Temperature    string `json:"temperature"`
	MaxContextTurn int    `json:"maxContextTurn"`
	EnableMultiAgent string `json:"enableMultiAgent"`
	DisplayThought   string `json:"displayThought"`
	DisplayPlugin    string `json:"displayPlugin"`

	SupportModelList  []AppModel  `json:"supportModelList"`
	AppPluginsInfoVos []AppPlugin `json:"appPluginsInfoVos"`
}

type AppModel struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

type AppPlugin struct {
	Name    string `json:"name"`
	Enabled string `json:"enabled"`
}

func NewCmdAppDetail(f *cmdutil.Factory, runF func(*DetailOptions) error) *cobra.Command {
	opts := &DetailOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "detail <code>",
		Short: "View application detail",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "app:read",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return detailRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")

	return cmd
}

func detailRun(opts *DetailOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("code", opts.Code)

	resp, err := client.Get(opts.Ctx, "/api/cli/app/detail", params)
	if err != nil {
		return fmt.Errorf("failed to get app detail: %w", err)
	}

	var detail AppDetail
	if err := resp.Decode(&detail); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, detail)
	}

	w := opts.Factory.IOStreams.Out

	fmt.Fprintf(w, "Name:        %s\n", detail.Name)
	fmt.Fprintf(w, "Code:        %s\n", detail.Code)
	fmt.Fprintf(w, "Type:        %s\n", detail.Type)
	if detail.AppStatus != "" {
		fmt.Fprintf(w, "Status:      %s\n", detail.AppStatus)
	}
	if detail.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", detail.Description)
	}
	if detail.Introduction != "" {
		fmt.Fprintf(w, "Introduction:%s\n", detail.Introduction)
	}
	if detail.NickName != "" {
		fmt.Fprintf(w, "Creator:     %s\n", detail.NickName)
	}
	fmt.Fprintf(w, "Usage:       %d  Likes: %d\n", detail.UsageCount, detail.ThumbCount)

	fmt.Fprintln(w, "\nConfig:")
	fmt.Fprintf(w, "  Temperature:       %s\n", detail.Temperature)
	fmt.Fprintf(w, "  Max context turns: %d\n", detail.MaxContextTurn)
	fmt.Fprintf(w, "  Multi-agent:       %s\n", yesNo(detail.EnableMultiAgent))
	fmt.Fprintf(w, "  Show thought:      %s\n", yesNo(detail.DisplayThought))
	fmt.Fprintf(w, "  Show plugin:       %s\n", yesNo(detail.DisplayPlugin))

	if len(detail.SupportModelList) > 0 {
		fmt.Fprintln(w, "\nModels:")
		for _, m := range detail.SupportModelList {
			tag := ""
			if m.IsDefault {
				tag = " (default)"
			}
			fmt.Fprintf(w, "  - %s%s\n", m.Name, tag)
		}
	}

	if len(detail.AppPluginsInfoVos) > 0 {
		fmt.Fprintln(w, "\nPlugins:")
		names := make([]string, 0, len(detail.AppPluginsInfoVos))
		for _, p := range detail.AppPluginsInfoVos {
			s := p.Name
			if p.Enabled == "N" {
				s += " (disabled)"
			}
			names = append(names, s)
		}
		fmt.Fprintf(w, "  %s\n", strings.Join(names, ", "))
	}

	return nil
}

func yesNo(v string) string {
	if v == "Y" {
		return "Yes"
	}
	return "No"
}
