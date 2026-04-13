package score

import (
	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

func NewCmdScore(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "score",
		Short: "Manage credits: view packages and purchase",
	}

	cmd.AddCommand(NewCmdScoreList(f, nil))
	cmd.AddCommand(NewCmdScoreBuy(f, nil))
	cmd.AddCommand(NewCmdScoreOrders(f, nil))
	cmd.AddCommand(NewCmdScoreOrder(f, nil))

	return cmd
}
