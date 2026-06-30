package account

import (
	"github.com/spf13/cobra"

	scoreCmd "github.com/MinimalFuture/linkai-cli/cmd/score"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

// NewCmdAccount creates the account command with subcommands.
//
// Credit/billing commands (formerly the top-level `score` command) are flattened
// under `account` since they belong to the same "account & recharge" concern:
//   - account info     → account profile, credits balance, plan
//   - account credits  → list purchasable credit packages
//   - account recharge → purchase credits
//   - account orders   → recharge order history
//   - account order    → single order status
func NewCmdAccount(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage account: profile, credits and recharge",
	}

	cmd.AddCommand(NewCmdAccountInfo(f, nil))
	cmd.AddCommand(scoreCmd.NewCmdScoreList(f, nil))   // account credits
	cmd.AddCommand(scoreCmd.NewCmdScoreBuy(f, nil))    // account recharge
	cmd.AddCommand(scoreCmd.NewCmdScoreOrders(f, nil)) // account orders
	cmd.AddCommand(scoreCmd.NewCmdScoreOrder(f, nil))  // account order

	return cmd
}
