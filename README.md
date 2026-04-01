# go-ratelimit

[![CI](https://github.com/philiprehberger/go-ratelimit/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-ratelimit/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-ratelimit.svg)](https://pkg.go.dev/github.com/philiprehberger/go-ratelimit)
[![Last updated](https://img.shields.io/github/last-commit/philiprehberger/go-ratelimit)](https://github.com/philiprehberger/go-ratelimit/commits/main)

Token bucket rate limiter for Go with per-key limiting and HTTP middleware. Zero external dependencies

## Installation

```bash
go get github.com/philiprehberger/go-ratelimit
```

## Usage

### Basic Limiter

```go
import "github.com/philiprehberger/go-ratelimit"

// Allow 10 requests per second with a burst of 20.
lim := ratelimit.New(10, 20)

if lim.Allow() {
    // Request allowed.
}

// Block until a token is available.
err := lim.Wait(ctx)
```

### Per-Key Limiter

```go
// Rate limit each user independently.
kl := ratelimit.NewKeyed(5, 10)

if kl.Allow(userID) {
    // Allowed for this user.
}

// Clean up inactive keys.
kl.Remove(userID)
```

### Runtime Configuration

```go
lim := ratelimit.New(10, 20)

// Later, adjust rate and burst without creating a new limiter.
lim.SetRate(50, 100)
```

### Metrics

```go
lim := ratelimit.New(10, 5)

lim.Allow() // true
lim.Allow() // true
lim.Allow() // false (if burst exhausted)

stats := lim.Stats()
fmt.Println(stats.Allowed)  // 2
fmt.Println(stats.Rejected) // 1
```

### Rejection Callback

```go
kl := ratelimit.NewKeyed(5, 10)

kl.OnReject(func(key string) {
    log.Printf("rate limited: %s", key)
})

kl.Allow(userID)
```

### HTTP Middleware

```go
// Global rate limit.
lim := ratelimit.New(100, 200)
mux.Handle("/api", ratelimit.Middleware(lim)(apiHandler))

// Per-IP rate limit.
kl := ratelimit.NewKeyed(10, 20)
mux.Handle("/api", ratelimit.KeyedMiddleware(kl, ratelimit.IPKeyFunc)(apiHandler))
```

## API

| Function / Type | Description |
|-----------------|-------------|
| `New(rate, burst)` | Create a token bucket limiter |
| `Limiter.Allow()` | Non-blocking check, returns true if token available |
| `Limiter.Wait(ctx)` | Block until token available or context cancelled |
| `Limiter.Tokens()` | Current available tokens |
| `Limiter.SetRate(rate, burst)` | Update rate and burst at runtime |
| `Limiter.Stats()` | Get allowed/rejected counters |
| `NewKeyed(rate, burst)` | Create a per-key limiter |
| `KeyedLimiter.Allow(key)` | Non-blocking per-key check |
| `KeyedLimiter.Wait(ctx, key)` | Blocking per-key wait |
| `KeyedLimiter.Remove(key)` | Remove a key's limiter |
| `KeyedLimiter.Size()` | Number of tracked keys |
| `KeyedLimiter.OnReject(fn)` | Register rejection callback |
| `Middleware(limiter)` | HTTP middleware returning 429 when exceeded |
| `KeyedMiddleware(limiter, keyFunc)` | Per-key HTTP middleware |
| `IPKeyFunc(r)` | Extract client IP as rate limit key |

## Development

```bash
go test ./...
go vet ./...
```

## Support

If you find this project useful:

⭐ [Star the repo](https://github.com/philiprehberger/go-ratelimit)

🐛 [Report issues](https://github.com/philiprehberger/go-ratelimit/issues?q=is%3Aissue+is%3Aopen+label%3Abug)

💡 [Suggest features](https://github.com/philiprehberger/go-ratelimit/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement)

❤️ [Sponsor development](https://github.com/sponsors/philiprehberger)

🌐 [All Open Source Projects](https://philiprehberger.com/open-source-packages)

💻 [GitHub Profile](https://github.com/philiprehberger)

🔗 [LinkedIn Profile](https://www.linkedin.com/in/philiprehberger)

## License

[MIT](LICENSE)
