package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestAllow_WithTokens(t *testing.T) {
	lim := New(10, 5)

	for i := 0; i < 5; i++ {
		if !lim.Allow() {
			t.Fatalf("Allow() = false on call %d, expected true (burst=5)", i+1)
		}
	}
}

func TestAllow_WithoutTokens(t *testing.T) {
	lim := New(10, 2)

	// Drain all tokens.
	lim.Allow()
	lim.Allow()

	if lim.Allow() {
		t.Fatal("Allow() = true after burst exhausted, expected false")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	lim := New(100, 1)

	if !lim.Allow() {
		t.Fatal("first Allow() should succeed")
	}
	if lim.Allow() {
		t.Fatal("second Allow() should fail (burst=1)")
	}

	// Wait enough time for a token to refill.
	time.Sleep(20 * time.Millisecond)

	if !lim.Allow() {
		t.Fatal("Allow() should succeed after refill time")
	}
}

func TestTokens_ReportsAvailable(t *testing.T) {
	lim := New(10, 5)

	tokens := lim.Tokens()
	if tokens < 4.9 || tokens > 5.1 {
		t.Fatalf("Tokens() = %f, expected ~5.0", tokens)
	}

	lim.Allow()
	tokens = lim.Tokens()
	if tokens < 3.9 || tokens > 4.1 {
		t.Fatalf("Tokens() after Allow() = %f, expected ~4.0", tokens)
	}
}

func TestWait_Success(t *testing.T) {
	lim := New(1000, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := lim.Wait(ctx); err != nil {
		t.Fatalf("Wait() error = %v, expected nil", err)
	}
}

func TestWait_WaitsForToken(t *testing.T) {
	lim := New(100, 1)

	// Drain the token.
	lim.Allow()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	if err := lim.Wait(ctx); err != nil {
		t.Fatalf("Wait() error = %v, expected nil", err)
	}
	elapsed := time.Since(start)

	if elapsed < 5*time.Millisecond {
		t.Fatalf("Wait() returned too quickly (%v), expected to wait for refill", elapsed)
	}
}

func TestWait_ContextCancelled(t *testing.T) {
	lim := New(0.001, 1) // Very slow rate.

	// Drain the token.
	lim.Allow()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := lim.Wait(ctx)
	if err == nil {
		t.Fatal("Wait() error = nil, expected context error")
	}
	if err != context.DeadlineExceeded {
		t.Fatalf("Wait() error = %v, expected context.DeadlineExceeded", err)
	}
}

func TestTokens_DoesNotExceedBurst(t *testing.T) {
	lim := New(1000, 3)

	// Wait a bit to let tokens accumulate.
	time.Sleep(10 * time.Millisecond)

	tokens := lim.Tokens()
	if tokens > 3.01 {
		t.Fatalf("Tokens() = %f, should not exceed burst of 3", tokens)
	}
}

func TestNew_StartsWithBurstTokens(t *testing.T) {
	lim := New(1, 10)

	tokens := lim.Tokens()
	if tokens < 9.9 || tokens > 10.1 {
		t.Fatalf("Tokens() = %f, expected ~10.0 at start", tokens)
	}
}

func TestZeroRate(t *testing.T) {
	lim := New(0, 2)

	// Should be able to use the initial burst tokens.
	if !lim.Allow() {
		t.Fatal("Allow() should succeed with initial burst")
	}
	if !lim.Allow() {
		t.Fatal("second Allow() should succeed with burst=2")
	}

	// No more tokens and rate is 0, so no refill.
	if lim.Allow() {
		t.Fatal("Allow() should fail with zero rate and no tokens")
	}
}
