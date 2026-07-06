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
	// Wait bounds how long a --device-code poll blocks, in seconds. It is only
	// meaningful together with --device-code (the agent polling path):
	//   <0  : not set → block until the code expires (interactive default)
	//    0  : single check — query once and return immediately
	//   >0  : poll for at most Wait seconds, then return pending
	Wait int
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
	cmd.Flags().IntVar(&opts.Wait, "wait", -1, "with --device-code: max seconds to poll before returning (0 = check once and return; >0 = bounded wait). Designed for AI-agent polling loops.")

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

	// --no-wait: return immediately with the URL + device code, plus an explicit
	// next-step instruction so an AI agent can drive the flow from this output
	// alone (without relying on the CLI skill prompt being loaded).
	if opts.NoWait {
		// Each poll blocks up to 60s so the user has plenty of time to open the
		// link, log in and approve on their phone, while still returning within
		// a single agent tool-call budget. The agent keeps re-running this until
		// authorization completes or the device code expires (expires_in).
		nextCommand := fmt.Sprintf("linkai auth login --device-code %s --wait 60 --json", authResp.DeviceCode)
		data := map[string]interface{}{
			"event":            "device_authorization",
			"verification_url": authResp.VerificationUriComplete,
			"device_code":      authResp.DeviceCode,
			"expires_in":       authResp.ExpiresIn,
			// Machine-readable guidance for an agent's next tool call.
			"next_action": map[string]interface{}{
				"instruction": "Show verification_url to the user and tell them to open it, log in and authorize (this may take a minute or two — scanning a QR code, logging in, etc.). Then, in this same turn (do not end your turn to wait for the user), run command as your next tool call to check status. Each call blocks up to 60s. It returns event=authorization_pending (NOT a failure — the user just hasn't finished yet; re-run the SAME command to keep waiting), authorization_complete (done), or authorization_failed (stop). Keep polling on 'pending' until complete or expires_in seconds have elapsed since this response.",
				"command":     nextCommand,
				"poll":        true,
			},
			"hint": fmt.Sprintf("Ask the user to open verification_url and authorize (may take 1-2 min), then poll (repeat on pending): %s", nextCommand),
		}
		return writeJSON(f, data)
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

// loginPollDeviceCode resumes polling with a device code from a previous
// --no-wait call.
//
// When --wait is set (>= 0) it uses the bounded poller intended for AI agents:
// it returns within the budget and reports one of three states as JSON —
// authorization_complete / authorization_pending / authorization_failed — so
// the agent can loop by calling this command again on "pending" without ever
// blocking a tool call indefinitely. Without --wait it blocks like the
// interactive flow.
func loginPollDeviceCode(opts *LoginOptions, cfg *config.Config) error {
	f := opts.Factory
	client := f.HttpClient()

	// Agent (bounded) mode: --wait explicitly provided.
	if opts.Wait >= 0 {
		result := auth.PollDeviceTokenBounded(opts.Ctx, client, cfg.APIBase(),
			opts.DeviceCode, 5, opts.Wait, f.IOStreams.ErrOut)

		if result.Pending {
			if opts.JSON {
				sameCommand := fmt.Sprintf("linkai auth login --device-code %s --wait %d --json", opts.DeviceCode, opts.Wait)
				return writeJSON(f, map[string]interface{}{
					"event":       "authorization_pending",
					"device_code": opts.DeviceCode,
					"next_action": map[string]interface{}{
						"instruction": "The user has not finished authorizing yet. This is NOT an error. Re-run the SAME command as your next tool call in this same turn (do not end your turn) to keep waiting. Stop only on authorization_complete or authorization_failed.",
						"command":     sameCommand,
						"poll":        true,
					},
					"hint": fmt.Sprintf("Not authorized yet — re-run to keep waiting: %s", sameCommand),
				})
			}
			fmt.Fprintln(f.IOStreams.ErrOut, "Authorization still pending; run again to keep waiting.")
			return nil
		}
		if !result.OK || result.Token == nil {
			if opts.JSON {
				_ = writeJSON(f, map[string]interface{}{
					"event": "authorization_failed",
					"error": result.Message,
				})
			}
			return fmt.Errorf("authorization failed: %s", result.Message)
		}
		return saveLoginResult(opts, cfg, result.Token)
	}

	// Interactive (blocking) mode.
	fmt.Fprintln(f.IOStreams.ErrOut, "Waiting for authorization...")
	result := auth.PollDeviceToken(opts.Ctx, client, cfg.APIBase(),
		opts.DeviceCode, 5, 300, f.IOStreams.ErrOut)

	if !result.OK {
		return fmt.Errorf("authorization failed: %s", result.Message)
	}
	if result.Token == nil {
		return fmt.Errorf("authorization succeeded but no token returned")
	}

	return saveLoginResult(opts, cfg, result.Token)
}

// writeJSON writes v as non-HTML-escaped JSON to stdout.
func writeJSON(f *cmdutil.Factory, v interface{}) error {
	enc := json.NewEncoder(f.IOStreams.Out)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
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
