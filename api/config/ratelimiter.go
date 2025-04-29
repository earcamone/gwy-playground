package config

import (
	"crypto/md5"
	"fmt"
	"github.com/earcamone/gapi/api/middleware/ratelimiter"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimiterKey allows client to specify which client request
// parameters wants to use to generate their rate limiting key
// At least one field must be set, or CustomPolicy will panic.
type RateLimiterKey struct {
	IP      bool
	Path    bool
	Method  bool
	Headers []string
}

// CustomPolicy creates a rate-limiting key generation function based
// on the provided RateLimiterKey.
//
// It inspects the http.Request to build a unique key from selected
// components (IP, Method, Path, Headers) in that order, then returns
// an MD5 hash of the result for consistent key length. If no components
// are selected or available, it falls back to the client IP.
//
// The function panics if the RateLimiterKey is empty.
//
// The client IP is extracted from common proxy headers (X-Forwarded-For,
// X-Real-IP, etc.), with ports stripped, falling back to RemoteAddr if no
// valid IP is found in headers.
//
// Examples:
//   - RateLimiterKey{IP: true} on a request with X-Forwarded-For: "203.0.113.1"
//     → Returns MD5("203.0.113.1")
//   - RateLimiterKey{Method: true, Path: true} on a POST to "/api/v1"
//     → Returns MD5("POST/api/v1")
//   - RateLimiterKey{IP: true, Headers: []string{"User-Agent"}} with IP "198.51.100.2" and User-Agent "curl"
//     → Returns MD5("198.51.100.2curl")
//   - RateLimiterKey{Headers: []string{"X-API-Key"}} with no X-API-Key header
//     → Returns MD5("127.0.0.1") (assuming RemoteAddr is "127.0.0.1:port")
func CustomPolicy(k RateLimiterKey) ratelimiter.RateLimiterPolicy {
	// Panic if the RateLimiterKey is completely empty
	if !k.IP && !k.Path && !k.Method && len(k.Headers) == 0 {
		panic("RateLimiterKey is empty; at least one field must be set")
	}

	return func(r *http.Request) string {
		var keyBuilder strings.Builder

		// Add IP if requested
		if k.IP {
			ip := getClientIP(r)
			keyBuilder.WriteString(ip)
		}

		// Add Method if requested
		if k.Method {
			keyBuilder.WriteString(r.Method)
		}

		// Add Path if requested
		if k.Path {
			keyBuilder.WriteString(r.URL.Path)
		}

		// Add Headers if specified
		if len(k.Headers) > 0 {
			for _, header := range k.Headers {
				if value := r.Header.Get(header); value != "" {
					keyBuilder.WriteString(value)
				}
			}
		}

		// If no components were added, fall back to client IP
		if keyBuilder.Len() == 0 {
			ip := getClientIP(r)
			return fmt.Sprintf("%x", md5.Sum([]byte(ip)))
		}

		// Compute MD5 hash of the combined key
		key := keyBuilder.String()
		return fmt.Sprintf("%x", md5.Sum([]byte(key)))
	}
}

// stripIPPort strip port from IP if present
// (e.g., "192.168.1.1:12345" -> "192.168.1.1")
func stripIPPort(ip string) string {
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}

	return ip
}

// getClientIP attempts to extract the real client IP from common proxy headers,
// falling back to RemoteAddr if no valid IP is found from headers.
// Ports are stripped from the result to ensure consistent key generation.
func getClientIP(r *http.Request) string {
	headers := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"CF-Connecting-IP",
		"True-Client-IP",
	}

	var ip string
	for _, header := range headers {
		if value := r.Header.Get(header); value != "" {
			if header == "X-Forwarded-For" {
				ips := strings.Split(value, ",")
				if len(ips) > 0 {
					ip = stripIPPort(strings.TrimSpace(ips[0]))
					if net.ParseIP(ip) != nil {
						break
					}
				}
			} else {
				ip = stripIPPort(strings.TrimSpace(value))
				if net.ParseIP(ip) != nil {
					break
				}
			}
		}
	}

	// If no valid IP from headers, use RemoteAddr as fallback
	if ip == "" || net.ParseIP(ip) == nil {
		ip = r.RemoteAddr
	}

	return ip
}

func TotalHitsPolicy(r *http.Request) string {
	return fmt.Sprintf("%x", md5.Sum([]byte("hits")))
}

func PerIPPolicy(r *http.Request) string {
	return CustomPolicy(RateLimiterKey{
		IP: true,
	})(r)
}

func PerPathPolicy(r *http.Request) string {
	return CustomPolicy(RateLimiterKey{
		Path: true,
	})(r)
}

func WithRateLimiter(policy ratelimiter.RateLimiterPolicy, hits uint64, window time.Duration) WithFunc {
	store := ratelimiter.NewMemoryStore(window)

	return func(config Config) Config {
		limiter := ratelimiter.RateLimiter(store, policy, hits)
		config.Middlewares = append(config.Middlewares, limiter)

		return config
	}
}

func WithDistributedRateLimiter(redis *redis.Client, policy ratelimiter.RateLimiterPolicy, hits uint64, window time.Duration) WithFunc {
	store := ratelimiter.NewRedisStore(redis, window)

	return func(config Config) Config {
		limiter := ratelimiter.RateLimiter(store, policy, hits)
		config.Middlewares = append(config.Middlewares, limiter)

		return config
	}
}
