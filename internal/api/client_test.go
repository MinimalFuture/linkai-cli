package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet_JSONEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":200,"message":"ok","data":{"name":"test"}}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), "tok")
	resp, err := c.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	var out struct{ Name string }
	if err := resp.Decode(&out); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out.Name != "test" {
		t.Errorf("Name = %q, want %q", out.Name, "test")
	}
}

func TestGet_NonJSONError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(502)
		w.Write([]byte(`<html><body>Bad Gateway</body></html>`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), "tok")
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for non-JSON 502 response")
	}
	// Should mention Content-Type, not "failed to parse response"
	if got := err.Error(); !contains(got, "text/html") {
		t.Errorf("error = %q, want mention of content type", got)
	}
}

func TestGet_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":403,"message":"forbidden","data":null}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), "tok")
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for API code 403")
	}
	if got := err.Error(); !contains(got, "forbidden") {
		t.Errorf("error = %q, want 'forbidden'", got)
	}
}

func TestStream_ErrorReadsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		w.Write([]byte(`{"message":"rate limited"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), "tok")
	_, err := c.Stream(context.Background(), "/test", map[string]string{})
	if err == nil {
		t.Fatal("expected error for 429 stream response")
	}
	if got := err.Error(); !contains(got, "rate limited") {
		t.Errorf("error = %q, want 'rate limited'", got)
	}
}

func TestIsJSONContentType(t *testing.T) {
	tests := []struct {
		ct   string
		want bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"text/json", true},
		{"text/html", false},
		{"", true},
	}
	for _, tt := range tests {
		if got := isJSONContentType(tt.ct); got != tt.want {
			t.Errorf("isJSONContentType(%q) = %v, want %v", tt.ct, got, tt.want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchSubstr(s, sub)
}

func searchSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
