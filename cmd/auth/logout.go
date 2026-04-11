package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/auth"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/config"
)

// LogoutOptions holds all inputs for auth logout.
type LogoutOptions struct {
	Factory *cmdutil.Factory
}

// NewCmdAuthLogout creates the auth logout subcommand.
func NewCmdAuthLogout(f *cmdutil.Factory, runF func(*LogoutOptions) error) *cobra.Command {
	opts := &LogoutOptions{Factory: f}

	return &cobra.Command{
		Use:   "logout",
		Short: "Log out (clear token)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return logoutRun(opts)
		},
	}
}

func logoutRun(opts *LogoutOptions) error {
	f := opts.Factory

	cfg, err := f.Config()
	if err != nil {
		return err
	}

	if cfg.User == nil {
		fmt.Fprintln(f.IOStreams.ErrOut, "Not logged in.")
		return nil
	}

	// Best-effort server-side token revocation so the token cannot be
	// reused even if it was leaked before this logout.
	if token := auth.GetStoredToken(); token != nil && token.AccessToken != "" {
		if err := auth.RevokeToken(f.HttpClient(), cfg.APIBase(), token.AccessToken); err != nil {
			fmt.Fprintf(f.IOStreams.ErrOut, "[linkai] [WARN] server-side revocation failed: %v\n", err)
		}
	}

	if err := auth.RemoveStoredToken(); err != nil {
		return fmt.Errorf("failed to remove token: %w", err)
	}

	cfg.User = nil
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintln(f.IOStreams.ErrOut, "✓ Logged out")
	return nil
}
