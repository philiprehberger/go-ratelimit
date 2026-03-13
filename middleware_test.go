package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_AllowsRequests(t *testing.T) {
	lim := New(10, 5)
	handler := Middleware(lim)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, expected %d", rec.Code, http.StatusOK)
	}
}

func TestMiddleware_Returns429WhenExceeded(t *testing.T) {
	lim := New(10, 1)
	handler := Middleware(lim)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request uses the burst token.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, expected %d", rec.Code, http.StatusOK)
	}

	// Second request should be rate limited.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: status = %d, expected %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestKeyedMiddleware_PerKeyLimiting(t *testing.T) {
	kl := NewKeyed(10, 1)
	handler := KeyedMiddleware(kl, IPKeyFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request from IP "1.2.3.4".
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first request from 1.2.3.4: status = %d, expected %d", rec.Code, http.StatusOK)
	}

	// Second request from same IP should be limited.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:12346"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request from 1.2.3.4: status = %d, expected %d", rec.Code, http.StatusTooManyRequests)
	}

	// Request from different IP should succeed.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.6.7.8:12345"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("request from 5.6.7.8: status = %d, expected %d", rec.Code, http.StatusOK)
	}
}

func TestKeyedMiddleware_CustomKeyFunc(t *testing.T) {
	kl := NewKeyed(10, 1)

	// Key by a custom header.
	keyFunc := func(r *http.Request) string {
		return r.Header.Get("X-API-Key")
	}

	handler := KeyedMiddleware(kl, keyFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request with API key "abc".
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "abc")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, expected %d", rec.Code, http.StatusOK)
	}

	// Second request with same key should be limited.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "abc")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: status = %d, expected %d", rec.Code, http.StatusTooManyRequests)
	}

	// Request with different key should succeed.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "def")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("different key request: status = %d, expected %d", rec.Code, http.StatusOK)
	}
}

func TestIPKeyFunc_ParsesHostPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:8080"

	key := IPKeyFunc(req)
	if key != "192.168.1.1" {
		t.Fatalf("IPKeyFunc() = %q, expected %q", key, "192.168.1.1")
	}
}

func TestIPKeyFunc_FallbackNoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1"

	key := IPKeyFunc(req)
	if key != "192.168.1.1" {
		t.Fatalf("IPKeyFunc() = %q, expected %q", key, "192.168.1.1")
	}
}

func TestMiddleware_DoesNotCallNextOn429(t *testing.T) {
	lim := New(10, 0) // Zero burst, no tokens available.
	called := false
	handler := Middleware(lim)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, expected %d", rec.Code, http.StatusTooManyRequests)
	}
	if called {
		t.Fatal("next handler was called despite rate limit exceeded")
	}
}
