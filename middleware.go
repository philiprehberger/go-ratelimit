package ratelimit

import (
	"net"
	"net/http"
)

// Middleware returns an HTTP middleware that rate limits all requests using
// the provided Limiter. When the limit is exceeded, it responds with
// 429 Too Many Requests.
func Middleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// KeyedMiddleware returns an HTTP middleware that rate limits requests on a
// per-key basis using the provided KeyedLimiter. The keyFunc extracts the
// rate limiting key from each request (e.g., client IP). When the limit for
// a key is exceeded, it responds with 429 Too Many Requests.
func KeyedMiddleware(limiter *KeyedLimiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if !limiter.Allow(key) {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IPKeyFunc extracts the client IP address from the request for use as a
// rate limiting key. It uses the host portion of RemoteAddr, stripping the port.
func IPKeyFunc(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
