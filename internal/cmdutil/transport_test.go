package cmdutil

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
)

type countTransport struct {
	calls    int32
	response func() *http.Response
}

func (t *countTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.AddInt32(&t.calls, 1)
	return t.response(), nil
}

func TestRetryTransport_NoRetryOnSuccess(t *testing.T) {
	inner := &countTransport{response: func() *http.Response {
		return &http.Response{StatusCode: 200, Body: http.NoBody}
	}}
	rt := newRetryTransport(inner, 3)
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&inner.calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

func TestRetryTransport_RetriesOn502(t *testing.T) {
	var attempt int32
	inner := &countTransport{response: func() *http.Response {
		n := atomic.AddInt32(&attempt, 1)
		if n <= 2 {
			return &http.Response{StatusCode: 502, Body: http.NoBody}
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}
	}}
	rt := newRetryTransport(inner, 3)
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&inner.calls); got != 3 {
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestRetryTransport_ReturnsLastResponseAfterExhaust(t *testing.T) {
	inner := &countTransport{response: func() *http.Response {
		return &http.Response{StatusCode: 503, Body: http.NoBody}
	}}
	rt := newRetryTransport(inner, 2)
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != 503 {
		t.Errorf("StatusCode = %d, want 503", resp.StatusCode)
	}
	// 1 initial + 2 retries = 3 total
	if got := atomic.LoadInt32(&inner.calls); got != 3 {
		t.Errorf("calls = %d, want 3", got)
	}
}
