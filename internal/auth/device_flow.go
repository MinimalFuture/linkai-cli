package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DeviceAuthResponse is the response from the device authorization endpoint.
type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	VerificationUri         string `json:"verification_uri"`
	VerificationUriComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// DeviceFlowTokenData contains the token data from a successful device flow.
type DeviceFlowTokenData struct {
	AccessToken      string
	RefreshToken     string
	Scope            string
	ExpiresIn        int // access token TTL in seconds
	RefreshExpiresIn int // refresh token TTL in seconds
	UserID           string
	UserName         string
	AccountNo        string
}

// DeviceFlowResult is the result of polling the token endpoint.
type DeviceFlowResult struct {
	OK      bool
	Token   *DeviceFlowTokenData
	Error   string
	Message string
}

// RequestDeviceAuthorization calls POST /api/cli/auth/device to initiate device flow.
// deviceID is the persistent per-machine device fingerprint.
// scope is the requested permission scope (e.g. "chat").
func RequestDeviceAuthorization(client *http.Client, apiBase, deviceID, scope string) (*DeviceAuthResponse, error) {
	url := apiBase + "/api/cli/auth/device"

	payload, _ := json.Marshal(map[string]string{
		"device_id": deviceID,
		"scope":     scope,
	})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device authorization request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("device authorization failed: HTTP %d – response not JSON", resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		msg := getString(raw, "message")
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("device authorization failed: %s", msg)
	}

	data, err := unwrapEnvelope(raw)
	if err != nil {
		return nil, fmt.Errorf("device authorization failed: %w", err)
	}

	if errStr := getString(data, "error"); errStr != "" {
		msg := getString(data, "error_description")
		if msg == "" {
			msg = errStr
		}
		return nil, fmt.Errorf("device authorization failed: %s", msg)
	}

	expiresIn := getInt(data, "expires_in", 300)
	interval := getInt(data, "interval", 3)

	verificationUri := getString(data, "verification_uri")
	verificationUriComplete := getString(data, "verification_uri_complete")
	if verificationUriComplete == "" {
		verificationUriComplete = verificationUri
	}

	return &DeviceAuthResponse{
		DeviceCode:              getString(data, "device_code"),
		VerificationUri:         verificationUri,
		VerificationUriComplete: verificationUriComplete,
		ExpiresIn:               expiresIn,
		Interval:                interval,
	}, nil
}

// PollDeviceToken polls POST /api/cli/auth/token until authorization completes or times out.
func PollDeviceToken(ctx context.Context, client *http.Client, apiBase, deviceCode string, interval, expiresIn int, errOut io.Writer) *DeviceFlowResult {
	if errOut == nil {
		errOut = io.Discard
	}

	const maxPollInterval = 60
	const maxPollAttempts = 200

	url := apiBase + "/api/cli/auth/token"
	deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)
	currentInterval := interval
	attempts := 0

	for time.Now().Before(deadline) && attempts < maxPollAttempts {
		attempts++
		if ctx.Err() != nil {
			return &DeviceFlowResult{OK: false, Error: "expired_token", Message: "Polling was cancelled"}
		}

		select {
		case <-time.After(time.Duration(currentInterval) * time.Second):
		case <-ctx.Done():
			return &DeviceFlowResult{OK: false, Error: "expired_token", Message: "Polling was cancelled"}
		}

		payload, _ := json.Marshal(map[string]string{
			"device_code": deviceCode,
		})
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(errOut, "[linkai] [WARN] device-flow: poll network error: %v\n", err)
			currentInterval = minInt(currentInterval+1, maxPollInterval)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Fprintf(errOut, "[linkai] [WARN] device-flow: poll read error: %v\n", err)
			currentInterval = minInt(currentInterval+1, maxPollInterval)
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			fmt.Fprintf(errOut, "[linkai] [WARN] device-flow: poll parse error: %v\n", err)
			currentInterval = minInt(currentInterval+1, maxPollInterval)
			continue
		}

		data, err := unwrapEnvelope(raw)
		if err != nil {
			fmt.Fprintf(errOut, "[linkai] [WARN] device-flow: poll envelope error: %v\n", err)
			return &DeviceFlowResult{OK: false, Error: "server_error", Message: err.Error()}
		}

		errStr := getString(data, "error")

		// Success: access_token present
		if errStr == "" && getString(data, "access_token") != "" {
			fmt.Fprintf(errOut, "[linkai] device-flow: token obtained successfully\n")
			return &DeviceFlowResult{
				OK: true,
				Token: &DeviceFlowTokenData{
					AccessToken:      getString(data, "access_token"),
					RefreshToken:     getString(data, "refresh_token"),
					Scope:            getString(data, "scope"),
					ExpiresIn:        getInt(data, "expires_in", 7200),
					RefreshExpiresIn: getInt(data, "refresh_expires_in", 7*24*3600),
					UserID:           getString(data, "user_id"),
					UserName:         getString(data, "user_name"),
					AccountNo:        getString(data, "account_no"),
				},
			}
		}

		switch errStr {
		case "authorization_pending":
			continue
		case "slow_down":
			currentInterval = minInt(currentInterval+5, maxPollInterval)
			fmt.Fprintf(errOut, "[linkai] device-flow: slow_down, interval increased to %ds\n", currentInterval)
			continue
		case "access_denied":
			msg := getString(data, "error_description")
			if msg == "" {
				msg = "Authorization denied by user"
			}
			return &DeviceFlowResult{OK: false, Error: "access_denied", Message: msg}
		case "expired_token", "invalid_grant":
			msg := getString(data, "error_description")
			if msg == "" {
				msg = "Device code expired, please try again"
			}
			return &DeviceFlowResult{OK: false, Error: "expired_token", Message: msg}
		}

		desc := getString(data, "error_description")
		if desc == "" {
			desc = errStr
		}
		if desc == "" {
			desc = "Unknown error"
		}
		fmt.Fprintf(errOut, "[linkai] [WARN] device-flow: unexpected error: %s\n", desc)
		return &DeviceFlowResult{OK: false, Error: "unexpected", Message: desc}
	}

	if attempts >= maxPollAttempts {
		fmt.Fprintf(errOut, "[linkai] [WARN] device-flow: max poll attempts reached\n")
	}
	return &DeviceFlowResult{OK: false, Error: "expired_token", Message: "Authorization timed out, please try again"}
}

// unwrapEnvelope checks if the raw JSON follows the ResultDTO envelope
// {"code":200,"msg":"...","data":{...}} and returns the inner data map.
// If the response is not wrapped, it returns the original map unchanged.
func unwrapEnvelope(m map[string]interface{}) (map[string]interface{}, error) {
	dataVal, hasData := m["data"]
	_, hasCode := m["code"]
	if !hasCode || !hasData {
		return m, nil
	}

	codeVal, _ := m["code"]
	var code int
	switch c := codeVal.(type) {
	case float64:
		code = int(c)
	case int:
		code = c
	default:
		return m, nil
	}

	if code != 200 {
		msg := getString(m, "message")
		if msg == "" {
			msg = fmt.Sprintf("server returned code %d", code)
		}
		return nil, fmt.Errorf("API error: %s", msg)
	}

	switch inner := dataVal.(type) {
	case map[string]interface{}:
		return inner, nil
	case nil:
		return m, nil
	default:
		return m, nil
	}
}

// helpers

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string, fallback int) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return fallback
}
