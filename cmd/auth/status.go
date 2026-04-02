package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	larkauth "github.com/yjr/linkai-cli/internal/auth"
	"github.com/yjr/linkai-cli/internal/cmdutil"
)

// StatusOptions holds all inputs for auth status.
type StatusOptions struct {
	Factory *cmdutil.Factory
	JSON    bool
}

// NewCmdAuthStatus creates the auth status subcommand.
func NewCmdAuthStatus(f *cmdutil.Factory, runF func(*StatusOptions) error) *cobra.Command {
	opts := &StatusOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "View current auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return statusRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")

	return cmd
}

func statusRun(opts *StatusOptions) error {
	f := opts.Factory

	cfg, err := f.Config()
	if err != nil {
		return err
	}

	result := map[string]interface{}{
		"api_base": cfg.APIBase(),
	}

	if cfg.User == nil {
		result["logged_in"] = false
		result["note"] = "Not logged in. Run 'linkai auth login' to authenticate."
		return printStatus(f, opts.JSON, result)
	}

	result["user_id"] = cfg.User.UserID
	result["user_name"] = cfg.User.UserName
	result["account_no"] = cfg.User.AccountNo

	stored := larkauth.GetStoredToken()
	if stored == nil {
		result["logged_in"] = false
		result["token_status"] = "missing"
		result["note"] = "Token not found. Run 'linkai auth login' to re-authenticate."
		return printStatus(f, opts.JSON, result)
	}

	status := larkauth.TokenStatus(stored)
	result["token_status"] = status
	result["granted_at"] = time.UnixMilli(stored.GrantedAt).Format(time.RFC3339)
	result["expires_at"] = time.UnixMilli(stored.ExpiresAt).Format(time.RFC3339)

	if stored.Scope != "" {
		result["scope"] = stored.Scope
	}
	if stored.RefreshExpiresAt > 0 {
		result["refresh_expires_at"] = time.UnixMilli(stored.RefreshExpiresAt).Format(time.RFC3339)
	}

	if status == "expired" {
		result["logged_in"] = false
		result["note"] = "Token has expired. Run 'linkai auth login' to re-authenticate."
	} else {
		result["logged_in"] = true
	}

	return printStatus(f, opts.JSON, result)
}

func printStatus(f *cmdutil.Factory, asJSON bool, result map[string]interface{}) error {
	if asJSON {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IOStreams.Out, string(data))
		return nil
	}

	// Human-readable output
	loggedIn, _ := result["logged_in"].(bool)
	if !loggedIn {
		fmt.Fprintln(f.IOStreams.Out, "Not logged in")
		fmt.Fprintln(f.IOStreams.Out, "Run 'linkai auth login' to authenticate")
		return nil
	}

	name, _ := result["user_name"].(string)
	account, _ := result["account_no"].(string)
	if name == "" {
		name = account
	}
	fmt.Fprintf(f.IOStreams.Out, "Logged in as %s (%s)\n", name, account)
	fmt.Fprintf(f.IOStreams.Out, "API:               %s\n", result["api_base"])

	if scope, ok := result["scope"].(string); ok && scope != "" {
		fmt.Fprintf(f.IOStreams.Out, "Scope:             %s\n", scope)
	}

	if tokenStatus, ok := result["token_status"].(string); ok {
		fmt.Fprintf(f.IOStreams.Out, "Token:             %s\n", tokenStatus)
		if grantedAt, ok := result["granted_at"].(string); ok {
			fmt.Fprintf(f.IOStreams.Out, "Granted at:        %s\n", grantedAt)
		}
		if expiresAt, ok := result["expires_at"].(string); ok {
			fmt.Fprintf(f.IOStreams.Out, "Expires at:        %s\n", expiresAt)
		}
		if refreshExp, ok := result["refresh_expires_at"].(string); ok {
			fmt.Fprintf(f.IOStreams.Out, "Refresh expires:   %s\n", refreshExp)
		}
	}

	if note, ok := result["note"].(string); ok {
		fmt.Fprintf(f.IOStreams.Out, "\n%s\n", note)
	}

	return nil
}
