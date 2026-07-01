package cmdutil

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/term"

	"github.com/MinimalFuture/linkai-cli/internal/api"
	"github.com/MinimalFuture/linkai-cli/internal/auth"
	"github.com/MinimalFuture/linkai-cli/internal/config"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

// ErrNotLoggedIn is returned by APIClient when no valid token is found.
var ErrNotLoggedIn = output.ErrAuth("not logged in: run 'linkai auth login'")

type Factory struct {
	Config           func() (*config.Config, error)
	HttpClient       func() *http.Client
	StreamHttpClient func() *http.Client // no timeout, for SSE / long-running requests
	IOStreams         *IOStreams
	APIClient        func() (*api.Client, error)
}

func NewDefault() *Factory {
	f := &Factory{}

	f.IOStreams = &IOStreams{
		In:              os.Stdin,
		Out:             os.Stdout,
		ErrOut:          os.Stderr,
		IsTerminal:      term.IsTerminal(int(os.Stdout.Fd())),
		IsStdinTerminal: term.IsTerminal(int(os.Stdin.Fd())),
	}

	var cachedConfig *config.Config
	f.Config = func() (*config.Config, error) {
		if cachedConfig != nil {
			return cachedConfig, nil
		}
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		cachedConfig = cfg
		return cachedConfig, nil
	}

	f.HttpClient = func() *http.Client {
		var deviceID string
		if cfg, err := f.Config(); err == nil {
			deviceID, _ = config.EnsureDeviceID(cfg)
		}
		return &http.Client{
			// Synchronous media generation (e.g. image gen) can take well over
			// 30s on some models, so allow a generous timeout. Per-request
			// cancellation is still honored via context.
			Timeout: 120 * time.Second,
			Transport: &deviceIDTransport{
				base:     newRetryTransport(http.DefaultTransport, 3),
				deviceID: deviceID,
			},
		}
	}

	f.StreamHttpClient = func() *http.Client {
		var deviceID string
		if cfg, err := f.Config(); err == nil {
			deviceID, _ = config.EnsureDeviceID(cfg)
		}
		return &http.Client{
			// No timeout — streaming connections are long-lived;
			// cancellation is handled via context.
			Transport: &deviceIDTransport{
				base:     newRetryTransport(http.DefaultTransport, 3),
				deviceID: deviceID,
			},
		}
	}

	var cachedAPIClient *api.Client
	f.APIClient = func() (*api.Client, error) {
		if cachedAPIClient != nil {
			return cachedAPIClient, nil
		}
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		token := auth.GetStoredToken()
		if token == nil {
			return nil, ErrNotLoggedIn
		}

		switch auth.TokenStatus(token) {
		case "expired":
			// The access token (2h TTL) has lapsed, which is expected for any
			// gap longer than that between commands. As long as the longer-
			// lived refresh token (7d TTL) is still valid, refresh instead of
			// forcing the user to re-login every couple of hours.
			now := time.Now().UnixMilli()
			if token.RefreshToken == "" || now >= token.RefreshExpiresAt {
				return nil, output.ErrAuth("token has expired: run 'linkai auth login'")
			}
			if err := refreshStoredToken(f, cfg, token); err != nil {
				return nil, output.ErrAuth(fmt.Sprintf("token has expired and refresh failed: %v — run 'linkai auth login'", err))
			}

		case "needs_refresh":
			if err := refreshStoredToken(f, cfg, token); err != nil {
				// Refresh failed — the current access token may still work for
				// a few more minutes, so fall through with it.
				fmt.Fprintf(f.IOStreams.ErrOut, "[linkai] [WARN] token refresh failed: %v\n", err)
			}
		}

		client := api.New(cfg.APIBase(), f.HttpClient(), token.AccessToken)
		client.StreamHTTPClient = f.StreamHttpClient()
		cachedAPIClient = client
		return cachedAPIClient, nil
	}

	return f
}

// refreshStoredToken exchanges the refresh token for a new access token and
// persists the result in place on token. Returns an error if the refresh
// call itself fails; token is left untouched in that case.
func refreshStoredToken(f *Factory, cfg *config.Config, token *auth.StoredToken) error {
	refreshed, err := auth.RefreshAccessToken(f.HttpClient(), cfg.APIBase(), token.RefreshToken)
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	token.AccessToken = refreshed.AccessToken
	token.ExpiresAt = now + int64(refreshed.ExpiresIn)*1000
	// The server rotates the refresh token only occasionally; on a plain
	// sliding-window refresh it keeps the same token value but extends its
	// TTL. Adopt a new token value when one is returned, and slide the local
	// expiry whenever the server reports a fresh refresh TTL.
	if refreshed.RefreshToken != "" {
		token.RefreshToken = refreshed.RefreshToken
	}
	if refreshed.RefreshExpiresIn > 0 {
		token.RefreshExpiresAt = now + int64(refreshed.RefreshExpiresIn)*1000
	}
	if writeErr := auth.SetStoredToken(token); writeErr != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "[linkai] [WARN] failed to persist refreshed token: %v\n", writeErr)
	}
	return nil
}

// deviceIDTransport injects X-Device-ID into every outgoing request.
type deviceIDTransport struct {
	base     http.RoundTripper
	deviceID string
}

func (t *deviceIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.deviceID != "" {
		req = req.Clone(req.Context())
		req.Header.Set("X-Device-ID", t.deviceID)
	}
	return t.base.RoundTrip(req)
}
