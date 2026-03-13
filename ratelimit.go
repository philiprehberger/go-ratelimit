// Package ratelimit provides token bucket rate limiting for Go.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter implements a token bucket rate limiter. Tokens are added at a fixed
// rate up to a maximum burst size. Each call to Allow or Wait consumes one token.
type Limiter struct {
	rate   float64
	burst  int
	tokens float64
	last   time.Time
	mu     sync.Mutex
}

// New creates a new Limiter that allows events at rate tokens per second with
// a maximum burst size of burst. The limiter starts with burst tokens available.
func New(rate float64, burst int) *Limiter {
	return &Limiter{
		rate:   rate,
		burst:  burst,
		tokens: float64(burst),
		last:   time.Now(),
	}
}

// Allow reports whether an event may happen now. It consumes one token if
// available and returns true. Otherwise it returns false without blocking.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// Wait blocks until a token is available or the context is cancelled.
// It returns nil when a token is consumed, or the context error if the
// context is cancelled before a token becomes available.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		l.refill()

		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}

		// Calculate how long until the next token is available.
		var wait time.Duration
		if l.rate > 0 {
			deficit := 1.0 - l.tokens
			wait = time.Duration(deficit / l.rate * float64(time.Second))
		} else {
			// Zero rate means tokens are never added; wait forever (context will cancel).
			wait = time.Hour
		}
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
			// Loop back to try again.
		}
	}
}

// Tokens returns the current number of available tokens. The value may be
// fractional because tokens are added continuously based on elapsed time.
func (l *Limiter) Tokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()
	return l.tokens
}

// refill adds tokens based on time elapsed since the last refill. The token
// count is capped at the burst size. Must be called while holding l.mu.
func (l *Limiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.last).Seconds()
	l.last = now

	l.tokens += elapsed * l.rate
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}
}
