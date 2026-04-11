package cmdutil

import (
	"math"
	"net/http"
	"time"
)

// retryTransport wraps an http.RoundTripper and retries on transient server
// errors (502, 503, 504) with exponential backoff.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
}

func newRetryTransport(base http.RoundTripper, maxRetries int) http.RoundTripper {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	return &retryTransport{base: base, maxRetries: maxRetries}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			select {
			case <-time.After(backoff):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		resp, err = t.base.RoundTrip(req)
		if err != nil {
			if req.Context().Err() != nil {
				return nil, err
			}
			continue
		}

		if resp.StatusCode != http.StatusBadGateway &&
			resp.StatusCode != http.StatusServiceUnavailable &&
			resp.StatusCode != http.StatusGatewayTimeout {
			return resp, nil
		}

		// Drain and close the body before retrying.
		resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}
	return resp, nil
}
