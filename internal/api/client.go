package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client is a LinkAI API client. It attaches Authorization and delegates
// X-Device-ID injection to the underlying http.Client transport (set up
// in cmdutil.Factory).
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
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
func (r *Response) Err() error {
	if r.OK() {
		return nil
	}
	return fmt.Errorf("API error %d: %s", r.Code, r.Msg)
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

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("stream request failed: HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// do executes the request, attaches auth, and decodes the standard envelope.
func (c *Client) do(req *http.Request) (*Response, error) {
	c.setAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var envelope Response
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse response (HTTP %d): %w", resp.StatusCode, err)
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
