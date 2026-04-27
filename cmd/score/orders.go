package score

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type Order struct {
	OrderNo    string    `json:"orderNo"`
	TotalFee   string    `json:"totalFee"`
	Status     string    `json:"status"`
	Score      int64     `json:"score"`
	BizType    string    `json:"bizType"`
	CreateTime time.Time `json:"createTime"`
}

type OrdersResult struct {
	Total    int64   `json:"total"`
	List     []Order `json:"list"`
	PageNum  int     `json:"pageNum"`
	PageSize int     `json:"pageSize"`
}

type OrdersOptions struct {
	Factory  *cmdutil.Factory
	Ctx      context.Context
	JSON     bool
	Page     int
	PageSize int
}

func NewCmdScoreOrders(f *cmdutil.Factory, runF func(*OrdersOptions) error) *cobra.Command {
	opts := &OrdersOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "orders",
		Short: "List credit purchase history",
		Annotations: map[string]string{
			permission.RequiredKey: permission.ScoreRead.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if runF != nil {
				return runF(opts)
			}
			return ordersRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "page number")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 10, "items per page")

	return cmd
}

func ordersRun(opts *OrdersOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("pageNo", strconv.Itoa(opts.Page))
	params.Set("pageSize", strconv.Itoa(opts.PageSize))

	resp, err := client.Get(opts.Ctx, "/api/cli/score/orders", params)
	if err != nil {
		return fmt.Errorf("failed to get orders: %w", err)
	}

	var result OrdersResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	if len(result.List) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No orders found.")
		return nil
	}

	headers := []string{"ORDER NO", "STATUS", "AMOUNT", "CREDITS", "DATE"}
	rows := make([][]string, 0, len(result.List))
	for _, o := range result.List {
		rows = append(rows, []string{
			o.OrderNo,
			o.Status,
			o.TotalFee,
			fmt.Sprintf("%d", o.Score),
			o.CreateTime.Format("2006-01-02 15:04"),
		})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)

	start := (opts.Page-1)*opts.PageSize + 1
	end := start + len(result.List) - 1
	fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\nShowing %d-%d of %d\n", start, end, result.Total)
	return nil
}
