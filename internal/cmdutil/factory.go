package cmdutil

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/term"

	"github.com/yjr/linkai-cli/internal/api"
	larkauth "github.com/yjr/linkai-cli/internal/auth"
	"github.com/yjr/linkai-cli/internal/config"
)

// ErrNotLoggedIn is returned by APIClient when no valid token is found.
var ErrNotLoggedIn = fmt.Errorf("not logged in: run 'linkai auth login'")

type Factory struct {
	Config     func() (*config.Config, error)
	HttpClient func() *http.Client
	IOStreams   *IOStreams
	APIClient  func() (*api.Client, error)
}

func NewDefault() *Factory {
	f := &Factory{}

	f.IOStreams = &IOStreams{
		In:         os.Stdin,
		Out:        os.Stdout,
		ErrOut:     os.Stderr,
		IsTerminal: term.IsTerminal(int(os.Stdout.Fd())),
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
			Timeout: 30 * time.Second,
			Transport: &deviceIDTransport{
				base:     http.DefaultTransport,
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
		token := larkauth.GetStoredToken()
		if token == nil {
			return nil, ErrNotLoggedIn
		}
		if larkauth.TokenStatus(token) == "expired" {
			return nil, fmt.Errorf("token has expired: run 'linkai auth login'")
		}
		cachedAPIClient = api.New(cfg.APIBase, f.HttpClient(), token.AccessToken)
		return cachedAPIClient, nil
	}

	return f
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
