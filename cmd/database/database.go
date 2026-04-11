package database

import (
	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

func NewCmdDatabase(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "database",
		Short: "Manage databases",
	}

	cmd.AddCommand(NewCmdDatabaseList(f, nil))
	cmd.AddCommand(NewCmdDatabaseTables(f, nil))
	cmd.AddCommand(NewCmdDatabaseDescribe(f, nil))
	cmd.AddCommand(NewCmdDatabaseExec(f, nil))

	return cmd
}
