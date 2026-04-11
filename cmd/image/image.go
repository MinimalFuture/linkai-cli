package image

import (
	"github.com/spf13/cobra"
	"github.com/yjr/linkai-cli/internal/cmdutil"
)

func NewCmdImage(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Generate images",
	}

	cmd.AddCommand(NewCmdImageGen(f, nil))

	return cmd
}
