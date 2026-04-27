package score

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type OrderOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	OrderNo string
}

func NewCmdScoreOrder(f *cmdutil.Factory, runF func(*OrderOptions) error) *cobra.Command {
	opts := &OrderOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "order <order-no>",
		Short: "Get order status by order number",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.ScoreRead.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.OrderNo = args[0]
			if runF != nil {
				return runF(opts)
			}
			return orderRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")

	return cmd
}

func orderRun(opts *OrderOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("orderNo", opts.OrderNo)

	resp, err := client.Get(opts.Ctx, "/api/cli/score/order/detail", params)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	var o Order
	if err := resp.Decode(&o); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, o)
	}

	w := opts.Factory.IOStreams.Out
	fmt.Fprintf(w, "Order No: %s\n", o.OrderNo)
	fmt.Fprintf(w, "Status:   %s\n", o.Status)
	fmt.Fprintf(w, "Amount:   %s\n", o.TotalFee)
	fmt.Fprintf(w, "Credits:  %d\n", o.Score)
	if !o.CreateTime.IsZero() {
		fmt.Fprintf(w, "Date:     %s\n", o.CreateTime.Format("2006-01-02 15:04:05"))
	}

	return nil
}
