package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/MinimalFuture/linkai-cli/internal/output"
)

// Client is a LinkAI API client. It attaches Authorization and delegates
// X-Device-ID injection to the underlying http.Client transport (set up
// in cmdutil.Factory).
type Client struct {
	BaseURL          string
	HTTPClient       *http.Client
	StreamHTTPClient *http.Client // optional; used for SSE / long-running requests
	Token            string
}

// New creates a new API client.
func New(baseURL string, httpClient *http.Client, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		Token:      token,
	}
}

// Response is the standard LinkAI API envelope.
type Response struct {
	Code int             `json:"code"`
	Msg  string          `json:"message"`
	Data json.RawMessage `json:"data"`
}

// OK returns true when the server reported success (code == 200).
func (r *Response) OK() bool { return r.Code == 200 }

// Decode unmarshals Data into v.
func (r *Response) Decode(v interface{}) error {
	return json.Unmarshal(r.Data, v)
}

// Err returns a non-nil error when the response code is not 200.
// It classifies errors by API code to produce appropriate exit codes.
func (r *Response) Err() error {
	if r.OK() {
		return nil
	}
	msg := fmt.Sprintf("API error %d: %s", r.Code, r.Msg)
	switch {
	case r.Code == 401 || r.Code == 403:
		return output.ErrAuth(msg)
	case r.Code >= 500:
		return output.ErrNetwork(msg)
	default:
		return output.Errorf("%s", msg)
	}
}

// Get sends a GET request to path with optional query parameters.
func (c *Client) Get(ctx context.Context, path string, params url.Values) (*Response, error) {
	u := c.BaseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// Post sends a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// Delete sends a DELETE request with optional query parameters.
func (c *Client) Delete(ctx context.Context, path string, params url.Values) (*Response, error) {
	u := c.BaseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// Stream opens an SSE / streaming response. The caller is responsible for
// closing the returned ReadCloser.
func (c *Client) Stream(ctx context.Context, path string, body interface{}) (io.ReadCloser, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	c.setAuth(req)

	httpClient := c.HTTPClient
	if c.StreamHTTPClient != nil {
		httpClient = c.StreamHTTPClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		detail := extractErrorDetail(errBody)
		if detail != "" {
			return nil, fmt.Errorf("stream request failed: HTTP %d – %s", resp.StatusCode, detail)
		}
		return nil, fmt.Errorf("stream request failed: HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// do executes the request, attaches auth, and decodes the standard envelope.
func (c *Client) do(req *http.Request) (*Response, error) {
	c.setAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, output.ErrNetwork("request failed: %v", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	ct := resp.Header.Get("Content-Type")
	if !isJSONContentType(ct) && resp.StatusCode >= 400 {
		snippet := string(raw)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "…"
		}
		return nil, fmt.Errorf("unexpected response (HTTP %d, Content-Type %q): %s", resp.StatusCode, ct, snippet)
	}

	var envelope Response
	if err := json.Unmarshal(raw, &envelope); err != nil {
		snippet := string(raw)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "…"
		}
		return nil, fmt.Errorf("failed to parse response (HTTP %d): body=%s", resp.StatusCode, snippet)
	}
	if !envelope.OK() {
		return nil, envelope.Err()
	}
	return &envelope, nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
}

// isJSONContentType returns true if ct looks like a JSON content type.
func isJSONContentType(ct string) bool {
	for _, s := range []string{"application/json", "text/json"} {
		if len(ct) >= len(s) && ct[:len(s)] == s {
			return true
		}
	}
	return ct == ""
}

// extractErrorDetail tries to pull a human-readable message from an error
// response body (JSON envelope or plain text).
func extractErrorDetail(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var obj map[string]interface{}
	if json.Unmarshal(body, &obj) == nil {
		if msg, ok := obj["message"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := obj["msg"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := obj["error"].(string); ok && msg != "" {
			return msg
		}
	}
	s := string(body)
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	return s
}
