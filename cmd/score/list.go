package score

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type Product struct {
	ID             int64  `json:"id"`
	ProductName    string `json:"productName"`
	Amount         string `json:"amount"`
	OriginalAmount string `json:"originalAmount"`
	AskCount       int    `json:"askCount"`
	ProductDesc    string `json:"productDesc"`
}

type ListOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
}

func NewCmdScoreList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "credits",
		Short: "List available credit packages",
		Annotations: map[string]string{
			permission.RequiredKey: permission.ScoreRead.String(),
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

	return cmd
}

func listRun(opts *ListOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(opts.Ctx, "/cli/score/products", nil)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}

	var products []Product
	if err := resp.Decode(&products); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, products)
	}

	if len(products) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No products available.")
		return nil
	}

	headers := []string{"ID", "NAME", "PRICE", "CREDITS", "DESCRIPTION"}
	rows := make([][]string, 0, len(products))
	for _, p := range products {
		desc := []rune(p.ProductDesc)
		if len(desc) > 30 {
			desc = append(desc[:30], []rune("...")...)
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", p.ID),
			p.ProductName,
			fmt.Sprintf("¥%s", p.Amount),
			fmt.Sprintf("%d", p.AskCount),
			string(desc),
		})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)
	return nil
}
