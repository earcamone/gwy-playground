package config

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomPolicyAndGetClientIP(t *testing.T) {
	newRequest := func(method, target string, body io.Reader) *http.Request {
		r := httptest.NewRequest(method, target, body)
		r.RemoteAddr = "127.0.0.1"

		return r
	}

	tests := []struct {
		name     string
		key      RateLimiterKey
		req      *http.Request
		expected string // Expected key string (MD5 computed in test)
	}{
		// Existing cases
		{
			name: "Only IP from X-Forwarded-For",
			key:  RateLimiterKey{IP: true},
			req: func() *http.Request {
				r := newRequest("GET", "/", nil)
				r.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")
				return r
			}(),
			expected: "203.0.113.1",
		},
		{
			name: "Only IP from X-Forwarded-For with Port",
			key:  RateLimiterKey{IP: true},
			req: func() *http.Request {
				r := newRequest("GET", "/", nil)
				r.Header.Set("X-Forwarded-For", "203.0.113.1:12345, 10.0.0.1")
				return r
			}(),
			expected: "203.0.113.1",
		},
		{
			name:     "Only IP from RemoteAddr",
			key:      RateLimiterKey{IP: true},
			req:      newRequest("GET", "/", nil),
			expected: "127.0.0.1",
		},
		{
			name: "IP from X-Real-IP with Port",
			key:  RateLimiterKey{IP: true},
			req: func() *http.Request {
				r := newRequest("GET", "/", nil)
				r.Header.Set("X-Real-IP", "198.51.100.2:8080")
				return r
			}(),
			expected: "198.51.100.2",
		},
		{
			name: "IP and Path",
			key:  RateLimiterKey{IP: true, Path: true},
			req: func() *http.Request {
				r := newRequest("GET", "/api/v1", nil)
				r.Header.Set("X-Real-IP", "198.51.100.2")
				return r
			}(),
			expected: "198.51.100.2/api/v1",
		},
		{
			name: "Method and Header",
			key: RateLimiterKey{
				Method:  true,
				Headers: []string{"User-Agent"},
			},
			req: func() *http.Request {
				r := newRequest("POST", "/", nil)
				r.Header.Set("User-Agent", "test-agent")
				return r
			}(),
			expected: "POSTtest-agent",
		},
		{
			name: "Headers Missing Fallback to IP",
			key: RateLimiterKey{
				Headers: []string{"X-Custom-Header"},
			},
			req:      newRequest("GET", "/", nil),
			expected: "127.0.0.1",
		},
		{
			name: "Multiple Headers",
			key: RateLimiterKey{
				Headers: []string{"User-Agent", "X-API-Key"},
			},
			req: func() *http.Request {
				r := newRequest("GET", "/", nil)
				r.Header.Set("User-Agent", "test-agent")
				r.Header.Set("X-API-Key", "abc123")
				return r
			}(),
			expected: "test-agentabc123",
		},
		{
			name: "Invalid Header Fallback to RemoteAddr",
			key:  RateLimiterKey{IP: true},
			req: func() *http.Request {
				r := newRequest("GET", "/", nil)
				r.Header.Set("X-Forwarded-For", "not-an-ip")
				return r
			}(),
			expected: "127.0.0.1",
		},
		// New cases from last round
		{
			name: "Invalid IP in X-Forwarded-For Multiple",
			key:  RateLimiterKey{IP: true},
			req: func() *http.Request {
				r := newRequest("GET", "/", nil)
				r.Header.Set("X-Forwarded-For", "not-an-ip, another-junk")
				return r
			}(),
			expected: "127.0.0.1",
		},
		{
			name: "IP from X-Real-IP without Port",
			key:  RateLimiterKey{IP: true},
			req: func() *http.Request {
				r := newRequest("GET", "/", nil)
				r.Header.Set("X-Real-IP", "198.51.100.2")
				return r
			}(),
			expected: "198.51.100.2",
		},
		// Single parameters
		{
			name:     "Only Method",
			key:      RateLimiterKey{Method: true},
			req:      newRequest("PUT", "/", nil),
			expected: "PUT",
		},
		{
			name:     "Only Path",
			key:      RateLimiterKey{Path: true},
			req:      newRequest("GET", "/users/123", nil),
			expected: "/users/123",
		},
		{
			name: "Only Header",
			key: RateLimiterKey{
				Headers: []string{"X-API-Key"},
			},
			req: func() *http.Request {
				r := newRequest("GET", "/", nil)
				r.Header.Set("X-API-Key", "xyz789")
				return r
			}(),
			expected: "xyz789",
		},
		// Two-parameter combinations
		{
			name: "IP and Method",
			key:  RateLimiterKey{IP: true, Method: true},
			req: func() *http.Request {
				r := newRequest("POST", "/", nil)
				r.Header.Set("X-Real-IP", "198.51.100.2")
				return r
			}(),
			expected: "198.51.100.2POST",
		},
		{
			name: "IP and Header",
			key: RateLimiterKey{
				IP:      true,
				Headers: []string{"User-Agent"},
			},
			req: func() *http.Request {
				r := newRequest("GET", "/", nil)
				r.Header.Set("X-Forwarded-For", "203.0.113.1")
				r.Header.Set("User-Agent", "test-agent")
				return r
			}(),
			expected: "203.0.113.1test-agent",
		},
		{
			name:     "Method and Path",
			key:      RateLimiterKey{Method: true, Path: true},
			req:      newRequest("DELETE", "/items/456", nil),
			expected: "DELETE/items/456",
		},
		{
			name: "Path and Header",
			key: RateLimiterKey{
				Path:    true,
				Headers: []string{"X-API-Key"},
			},
			req: func() *http.Request {
				r := newRequest("GET", "/api/v2", nil)
				r.Header.Set("X-API-Key", "abc123")
				return r
			}(),
			expected: "/api/v2abc123",
		},
		// Three-parameter combinations
		{
			name: "IP, Method, Path",
			key:  RateLimiterKey{IP: true, Method: true, Path: true},
			req: func() *http.Request {
				r := newRequest("PATCH", "/data", nil)
				r.Header.Set("X-Real-IP", "198.51.100.2")
				return r
			}(),
			expected: "198.51.100.2PATCH/data",
		},
		{
			name: "IP, Method, Header",
			key: RateLimiterKey{
				IP:      true,
				Method:  true,
				Headers: []string{"User-Agent"},
			},
			req: func() *http.Request {
				r := newRequest("POST", "/", nil)
				r.Header.Set("X-Forwarded-For", "203.0.113.1")
				r.Header.Set("User-Agent", "test-agent")
				return r
			}(),
			expected: "203.0.113.1POSTtest-agent",
		},
		{
			name: "IP, Path, Header",
			key: RateLimiterKey{
				IP:      true,
				Path:    true,
				Headers: []string{"X-API-Key"},
			},
			req: func() *http.Request {
				r := newRequest("GET", "/api/v3", nil)
				r.Header.Set("X-Real-IP", "198.51.100.2")
				r.Header.Set("X-API-Key", "xyz789")
				return r
			}(),
			expected: "198.51.100.2/api/v3xyz789",
		},
		{
			name: "Method, Path, Header",
			key: RateLimiterKey{
				Method:  true,
				Path:    true,
				Headers: []string{"User-Agent"},
			},
			req: func() *http.Request {
				r := newRequest("PUT", "/users/789", nil)
				r.Header.Set("User-Agent", "test-agent")
				return r
			}(),
			expected: "PUT/users/789test-agent",
		},
		// All four parameters
		{
			name: "IP, Method, Path, Header",
			key: RateLimiterKey{
				IP:      true,
				Method:  true,
				Path:    true,
				Headers: []string{"X-API-Key"},
			},
			req: func() *http.Request {
				r := newRequest("DELETE", "/items/999", nil)
				r.Header.Set("X-Forwarded-For", "203.0.113.1")
				r.Header.Set("X-API-Key", "abc123")
				return r
			}(),
			expected: "203.0.113.1DELETE/items/999abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := CustomPolicy(tt.key)
			result := policy(tt.req)
			expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(tt.expected)))
			if result != expectedHash {
				t.Errorf("expected %q (hash of %q), got %q", expectedHash, tt.expected, result)
			}
		})
	}
}

func TestCustomPolicyPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty RateLimiterKey")
		}
	}()
	CustomPolicy(RateLimiterKey{})
}

func TestBuiltInPolicies(t *testing.T) {
	tests := []struct {
		name     string
		policy   func(r *http.Request) string
		req      *http.Request
		expected string
	}{
		{
			name:   "TotalHitsPolicy",
			policy: TotalHitsPolicy,
			req: func() *http.Request {
				r := httptest.NewRequest("GET", "/test", nil)
				r.RemoteAddr = "127.0.0.1:12345"
				return r
			}(),
			expected: "hits",
		},
		{
			name:   "PerIPPolicy",
			policy: PerIPPolicy,
			req: func() *http.Request {
				r := httptest.NewRequest("GET", "/test", nil)
				r.RemoteAddr = "127.0.0.1:12345"
				r.Header.Set("X-Forwarded-For", "203.0.113.1")
				return r
			}(),
			expected: "203.0.113.1",
		},
		{
			name:   "PerPathPolicy",
			policy: PerPathPolicy,
			req: func() *http.Request {
				r := httptest.NewRequest("GET", "/api/v1", nil)
				r.RemoteAddr = "127.0.0.1:12345"
				return r
			}(),
			expected: "/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.policy(tt.req)
			expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(tt.expected)))
			if result != expectedHash {
				t.Errorf("expected %q (hash of %q), got %q", expectedHash, tt.expected, result)
			}
		})
	}
}
