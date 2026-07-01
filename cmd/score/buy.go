package score

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/api"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

const (
	pendingOrderFile = "pending_order.json"
	pendingOrderTTL  = 30 * time.Minute
	pollInterval     = 3 * time.Second
	pollTimeout      = 10 * time.Minute
)

type pendingOrder struct {
	OrderNo    string `json:"order_no"`
	ProductID  string `json:"product_id"`
	PayChannel string `json:"pay_channel"`
	CodeURL    string `json:"code_url"`
	CreatedAt  int64  `json:"created_at"`
}

type OrderCreateResult struct {
	OrderNo   string `json:"orderNo"`
	CodeURL   string `json:"codeUrl"`
	URLBase64 string `json:"urlBase64"`
}

type OrderDetail struct {
	OrderNo  string `json:"orderNo"`
	Status   string `json:"status"`
	Score    int64  `json:"score"`
	TotalFee string `json:"totalFee"`
}

type BuyOptions struct {
	Factory    *cmdutil.Factory
	Ctx        context.Context
	JSON       bool
	Agent      bool
	ProductID  string
	PayChannel string
}

func NewCmdScoreBuy(f *cmdutil.Factory, runF func(*BuyOptions) error) *cobra.Command {
	opts := &BuyOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "recharge",
		Short: "Recharge credits (purchase)",
		Annotations: map[string]string{
			permission.RequiredKey: permission.ScoreBuy.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if opts.JSON || !opts.Factory.IOStreams.IsTerminal {
				opts.Agent = true
			}
			if runF != nil {
				return runF(opts)
			}
			return buyRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format (agent mode)")
	cmd.Flags().BoolVar(&opts.Agent, "agent", false, "agent mode: return QR code URL instead of ASCII QR")
	cmd.Flags().StringVar(&opts.ProductID, "product", "", "product ID (skip interactive selection)")
	cmd.Flags().StringVar(&opts.PayChannel, "pay", "wechat", "payment channel: wechat or alipay")

	return cmd
}

func buyRun(opts *BuyOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	// Step 1: resolve product ID
	productID := opts.ProductID
	if productID == "" {
		if opts.Agent {
			return fmt.Errorf("--product is required in agent/JSON mode")
		}
		productID, err = selectProduct(opts, client)
		if err != nil {
			return err
		}
	}

	// Step 2: get or create order (with dedup)
	orderNo, codeURL, err := resolveOrder(opts, client, productID)
	if err != nil {
		return err
	}

	// Step 3: output
	if opts.Agent {
		return agentOutput(opts, orderNo, codeURL, productID)
	}
	return humanFlow(opts, client, orderNo, codeURL)
}

func selectProduct(opts *BuyOptions, client *api.Client) (string, error) {
	resp, err := client.Get(opts.Ctx, "/cli/score/products", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get products: %w", err)
	}

	var products []Product
	if err := resp.Decode(&products); err != nil {
		return "", fmt.Errorf("failed to parse products: %w", err)
	}
	if len(products) == 0 {
		return "", fmt.Errorf("no products available")
	}

	out := opts.Factory.IOStreams.Out
	fmt.Fprintln(out, "Available credit packages:")
	for i, p := range products {
		fmt.Fprintf(out, "  %d) %-20s ¥%-8s  %d credits\n", i+1, p.ProductName, p.Amount, p.AskCount)
	}

	fmt.Fprintf(out, "\nSelect package [1-%d]: ", len(products))
	var choice int
	if _, err := fmt.Fscan(opts.Factory.IOStreams.In, &choice); err != nil || choice < 1 || choice > len(products) {
		return "", fmt.Errorf("invalid selection")
	}
	return fmt.Sprintf("%d", products[choice-1].ID), nil
}

func resolveOrder(opts *BuyOptions, client *api.Client, productID string) (string, string, error) {
	// Check local pending order cache
	if po, err := loadPendingOrder(); err == nil && po != nil {
		cacheMatch := po.ProductID == productID &&
			po.PayChannel == opts.PayChannel &&
			time.Since(time.Unix(po.CreatedAt, 0)) < pendingOrderTTL
		if cacheMatch {
			params := url.Values{}
			params.Set("orderNo", po.OrderNo)
			resp, err := client.Get(opts.Ctx, "/cli/score/order/detail", params)
			if err == nil {
				var detail OrderDetail
				if decErr := resp.Decode(&detail); decErr == nil {
					switch detail.Status {
					case "INIT":
						fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "Reusing existing pending order %s\n", po.OrderNo)
						return po.OrderNo, po.CodeURL, nil
					case "PAID", "FAILED", "REFUND", "CANCELED":
						// Terminal state: safe to clear the cache and create a new order
						_ = removePendingOrder()
					}
					// Unknown status: fall through to create new order without clearing cache
				}
				// Decode error: fall through without clearing cache
			}
			// Network/API error: fall through without clearing cache to avoid losing a still-valid order
		} else {
			// Different product/channel or TTL expired: clear stale cache
			_ = removePendingOrder()
		}
	}

	// Create new order
	postResp, err := client.Post(opts.Ctx, "/cli/score/order/create", map[string]string{
		"goodId":     productID,
		"payChannel": opts.PayChannel,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create order: %w", err)
	}

	var created OrderCreateResult
	if err := postResp.Decode(&created); err != nil {
		return "", "", fmt.Errorf("failed to parse order response: %w", err)
	}

	_ = savePendingOrder(&pendingOrder{
		OrderNo:    created.OrderNo,
		ProductID:  productID,
		PayChannel: opts.PayChannel,
		CodeURL:    created.CodeURL,
		CreatedAt:  time.Now().Unix(),
	})

	return created.OrderNo, created.CodeURL, nil
}

func agentOutput(opts *BuyOptions, orderNo, codeURL, productID string) error {
	result := map[string]string{
		"order_no":    orderNo,
		"product_id":  productID,
		"pay_channel": opts.PayChannel,
		"code_url":    codeURL,
		"status":      "INIT",
	}
	if codeURL != "" {
		if png, err := qrcode.Encode(codeURL, qrcode.Medium, 300); err == nil {
			result["qr_base64"] = base64.StdEncoding.EncodeToString(png)
		}
	}
	return output.PrintJSON(opts.Factory.IOStreams.Out, result)
}

func humanFlow(opts *BuyOptions, client *api.Client, orderNo, codeURL string) error {
	out := opts.Factory.IOStreams.Out
	errOut := opts.Factory.IOStreams.ErrOut

	fmt.Fprintf(out, "\nOrder: %s\nScan the QR code to pay:\n\n", orderNo)

	q, err := qrcode.New(codeURL, qrcode.Medium)
	if err != nil {
		fmt.Fprintf(out, "QR code URL: %s\n", codeURL)
	} else {
		fmt.Fprintln(out, q.ToString(false))
	}

	fmt.Fprintln(errOut, "Waiting for payment...")

	deadline := time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-time.After(pollInterval):
		case <-opts.Ctx.Done():
			return fmt.Errorf("canceled — order %s still pending", orderNo)
		}

		params := url.Values{}
		params.Set("orderNo", orderNo)
		resp, err := client.Get(opts.Ctx, "/cli/score/order/detail", params)
		if err != nil {
			continue
		}

		var detail OrderDetail
		if err := resp.Decode(&detail); err != nil {
			continue
		}

		switch detail.Status {
		case "PAID":
			_ = removePendingOrder()
			fmt.Fprintf(out, "\nPayment successful! %d credits added to your account.\n", detail.Score)
			return nil
		case "FAILED", "REFUND":
			_ = removePendingOrder()
			return fmt.Errorf("payment %s", detail.Status)
		}
	}

	return fmt.Errorf("payment timeout — order %s still pending, retry to reuse it", orderNo)
}

// ── pending order helpers ──────────────────────────────────────────────────

func pendingOrderPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".linkai", pendingOrderFile)
}

func loadPendingOrder() (*pendingOrder, error) {
	data, err := os.ReadFile(pendingOrderPath())
	if err != nil {
		return nil, err
	}
	var po pendingOrder
	if err := json.Unmarshal(data, &po); err != nil {
		return nil, err
	}
	return &po, nil
}

func savePendingOrder(po *pendingOrder) error {
	data, err := json.Marshal(po)
	if err != nil {
		return err
	}
	return os.WriteFile(pendingOrderPath(), data, 0600)
}

func removePendingOrder() error {
	return os.Remove(pendingOrderPath())
}
