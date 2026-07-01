package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/auth"
	"github.com/MinimalFuture/linkai-cli/internal/browser"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/config"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

// LoginOptions holds all inputs for auth login.
type LoginOptions struct {
	Factory    *cmdutil.Factory
	Ctx        context.Context
	JSON       bool
	NoWait     bool
	DeviceCode string
	Scope      string
}

// NewCmdAuthLogin creates the auth login subcommand.
func NewCmdAuthLogin(f *cmdutil.Factory, runF func(*LoginOptions) error) *cobra.Command {
	opts := &LoginOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to LinkAI platform via browser authorization",
		Long: `Log in to LinkAI platform via browser authorization.

Opens a verification URL — open it in your browser to authorize the CLI.
This command blocks until authorization is complete.

For AI agents: run this command in the background and retrieve the
verification URL from its output. Use --no-wait to get the URL immediately.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if runF != nil {
				return runF(opts)
			}
			return loginRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "structured JSON output")
	cmd.Flags().BoolVar(&opts.NoWait, "no-wait", false, "initiate device authorization and return immediately; use --device-code to complete")
	cmd.Flags().StringVar(&opts.DeviceCode, "device-code", "", "poll and complete authorization with a device code from a previous --no-wait call")
	cmd.Flags().StringVar(&opts.Scope, "scope", permission.Defaults(), "space-separated list of requested permission scopes")

	return cmd
}

func loginRun(opts *LoginOptions) error {
	f := opts.Factory

	cfg, err := f.Config()
	if err != nil {
		return err
	}

	log := func(format string, a ...interface{}) {
		if !opts.JSON {
			fmt.Fprintf(f.IOStreams.ErrOut, format+"\n", a...)
		}
	}

	// --device-code: resume polling from a previous --no-wait call
	if opts.DeviceCode != "" {
		return loginPollDeviceCode(opts, cfg)
	}

	// Already logged in? Skip Device Flow unless the user is requesting new scopes.
	if existing := auth.GetStoredToken(); existing != nil {
		status := auth.TokenStatus(existing)
		if (status == "valid" || status == "needs_refresh") && permission.Covered(opts.Scope, existing.Scope) {
			userName := ""
			if cfg.User != nil {
				userName = cfg.User.UserName
			}
			if opts.JSON {
				data := map[string]interface{}{
					"event":     "already_logged_in",
					"user_name": userName,
					"scope":     existing.Scope,
				}
				enc := json.NewEncoder(f.IOStreams.Out)
				enc.SetEscapeHTML(false)
				return enc.Encode(data)
			}
			if userName != "" {
				fmt.Fprintf(f.IOStreams.ErrOut, "Already logged in as %s (scope: %s). Use `linkai auth logout` first to switch accounts.\n", userName, existing.Scope)
			} else {
				fmt.Fprintf(f.IOStreams.ErrOut, "Already logged in (scope: %s). Use `linkai auth logout` first to switch accounts.\n", existing.Scope)
			}
			return nil
		}
	}

	// Get persistent device ID from config
	deviceID, err := config.EnsureDeviceID(cfg)
	if err != nil {
		return fmt.Errorf("failed to get device ID: %w", err)
	}

	scope := opts.Scope
	if scope == "" {
		scope = "chat"
	}

	// Step 1: Request device authorization
	client := f.HttpClient()
	authResp, err := auth.RequestDeviceAuthorization(client, cfg.APIBase(), deviceID, scope)
	if err != nil {
		return fmt.Errorf("device authorization failed: %w", err)
	}

	// --no-wait: return immediately with device code and URL
	if opts.NoWait {
		data := map[string]interface{}{
			"verification_url": authResp.VerificationUriComplete,
			"device_code":      authResp.DeviceCode,
			"expires_in":       authResp.ExpiresIn,
			"hint":             fmt.Sprintf("Open verification_url in browser, then run: linkai auth login --device-code %s", authResp.DeviceCode),
		}
		enc := json.NewEncoder(f.IOStreams.Out)
		enc.SetEscapeHTML(false)
		return enc.Encode(data)
	}

	// The verification URL is returned fully-formed by the server (configured
	// per-environment via cli.verification.uri), and may live on a different
	// host/port than the API (e.g. frontend :9090 vs API :8901 in local dev).
	// Use it verbatim — do not rebase it onto the API base.

	// Step 2: Show verification URL
	if opts.JSON {
		data := map[string]interface{}{
			"event":                     "device_authorization",
			"verification_uri":          authResp.VerificationUri,
			"verification_uri_complete": authResp.VerificationUriComplete,
			"expires_in":                authResp.ExpiresIn,
		}
		enc := json.NewEncoder(f.IOStreams.Out)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(data)
	} else {
		fmt.Fprintf(f.IOStreams.ErrOut, "\nRequesting authorization with scope: %s\n", scope)
		fmt.Fprintf(f.IOStreams.ErrOut, "Open the following URL in your browser to authorize:\n")
		fmt.Fprintf(f.IOStreams.ErrOut, "  %s\n\n", authResp.VerificationUriComplete)
		// Best-effort: try to pop the browser open automatically. On headless
		// hosts (containers, remote Linux without a display) this fails
		// silently — the URL above is always printed for manual open.
		if err := browser.Open(authResp.VerificationUriComplete); err == nil {
			fmt.Fprintf(f.IOStreams.ErrOut, "(opened in your default browser)\n\n")
		}
	}

	// Step 3: Poll for token
	log("Waiting for authorization...")
	result := auth.PollDeviceToken(opts.Ctx, client, cfg.APIBase(),
		authResp.DeviceCode, authResp.Interval, authResp.ExpiresIn, f.IOStreams.ErrOut)

	if !result.OK {
		if opts.JSON {
			data := map[string]interface{}{
				"event": "authorization_failed",
				"error": result.Message,
			}
			enc := json.NewEncoder(f.IOStreams.Out)
			enc.SetEscapeHTML(false)
			_ = enc.Encode(data)
		}
		return fmt.Errorf("authorization failed: %s", result.Message)
	}

	return saveLoginResult(opts, cfg, result.Token)
}

// loginPollDeviceCode resumes polling with a device code from a previous --no-wait call.
func loginPollDeviceCode(opts *LoginOptions, cfg *config.Config) error {
	f := opts.Factory
	client := f.HttpClient()

	fmt.Fprintln(f.IOStreams.ErrOut, "Waiting for authorization...")
	result := auth.PollDeviceToken(opts.Ctx, client, cfg.APIBase(),
		opts.DeviceCode, 3, 300, f.IOStreams.ErrOut)

	if !result.OK {
		return fmt.Errorf("authorization failed: %s", result.Message)
	}
	if result.Token == nil {
		return fmt.Errorf("authorization succeeded but no token returned")
	}

	return saveLoginResult(opts, cfg, result.Token)
}

// saveLoginResult stores the token and user info after successful authorization.
func saveLoginResult(opts *LoginOptions, cfg *config.Config, token *auth.DeviceFlowTokenData) error {
	f := opts.Factory

	now := time.Now().UnixMilli()
	storedToken := &auth.StoredToken{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		Scope:            token.Scope,
		ExpiresAt:        now + int64(token.ExpiresIn)*1000,
		RefreshExpiresAt: now + int64(token.RefreshExpiresIn)*1000,
		GrantedAt:        now,
	}
	if err := auth.SetStoredToken(storedToken); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	cfg.User = &config.AppUser{
		UserName: token.UserName,
	}
	if err := config.Save(cfg); err != nil {
		// Rollback: remove the token we just wrote to avoid split-brain state
		_ = auth.RemoveStoredToken()
		return fmt.Errorf("failed to save config: %w", err)
	}

	if opts.JSON {
		data := map[string]interface{}{
			"event":     "authorization_complete",
			"user_name": token.UserName,
			"scope":     token.Scope,
		}
		enc := json.NewEncoder(f.IOStreams.Out)
		enc.SetEscapeHTML(false)
		return enc.Encode(data)
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "\n✓ Logged in as %s\n", token.UserName)
	return nil
}
