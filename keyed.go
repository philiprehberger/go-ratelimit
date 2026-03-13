package ratelimit

import (
	"context"
	"sync"
)

// KeyedLimiter manages per-key rate limiters. Each unique key gets its own
// independent Limiter instance, useful for per-user or per-IP rate limiting.
type KeyedLimiter struct {
	limiters map[string]*Limiter
	mu       sync.Mutex
	rate     float64
	burst    int
}

// NewKeyed creates a new KeyedLimiter. Each key that is seen for the first time
// will get a new Limiter configured with the given rate and burst.
func NewKeyed(rate float64, burst int) *KeyedLimiter {
	return &KeyedLimiter{
		limiters: make(map[string]*Limiter),
		rate:     rate,
		burst:    burst,
	}
}

// Allow reports whether an event for the given key may happen now. It returns
// true if a token is available for that key, consuming it. If the key has not
// been seen before, a new limiter is created automatically.
func (kl *KeyedLimiter) Allow(key string) bool {
	return kl.getLimiter(key).Allow()
}

// Wait blocks until a token is available for the given key or the context is
// cancelled. If the key has not been seen before, a new limiter is created.
func (kl *KeyedLimiter) Wait(ctx context.Context, key string) error {
	return kl.getLimiter(key).Wait(ctx)
}

// Remove deletes the limiter for the given key. Future calls with this key
// will create a fresh limiter.
func (kl *KeyedLimiter) Remove(key string) {
	kl.mu.Lock()
	defer kl.mu.Unlock()
	delete(kl.limiters, key)
}

// Size returns the number of keys currently being tracked.
func (kl *KeyedLimiter) Size() int {
	kl.mu.Lock()
	defer kl.mu.Unlock()
	return len(kl.limiters)
}

// getLimiter returns the limiter for the given key, creating one if it does
// not exist.
func (kl *KeyedLimiter) getLimiter(key string) *Limiter {
	kl.mu.Lock()
	defer kl.mu.Unlock()

	lim, ok := kl.limiters[key]
	if !ok {
		lim = New(kl.rate, kl.burst)
		kl.limiters[key] = lim
	}
	return lim
}
