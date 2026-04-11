package model

import (
	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

func NewCmdModel(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Manage AI models",
	}

	cmd.AddCommand(NewCmdModelList(f, nil))

	return cmd
}
