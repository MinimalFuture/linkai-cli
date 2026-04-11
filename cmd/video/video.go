package video

import (
	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

func NewCmdVideo(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Generate videos",
	}

	cmd.AddCommand(NewCmdVideoGen(f, nil))

	return cmd
}
