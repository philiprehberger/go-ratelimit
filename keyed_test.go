package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestKeyedLimiter_SeparateKeys(t *testing.T) {
	kl := NewKeyed(10, 1)

	if !kl.Allow("a") {
		t.Fatal("Allow(a) should succeed")
	}
	if !kl.Allow("b") {
		t.Fatal("Allow(b) should succeed (separate key)")
	}

	// Key "a" should be exhausted.
	if kl.Allow("a") {
		t.Fatal("Allow(a) should fail after burst exhausted")
	}

	// Key "b" should also be exhausted.
	if kl.Allow("b") {
		t.Fatal("Allow(b) should fail after burst exhausted")
	}
}

func TestKeyedLimiter_Allow(t *testing.T) {
	kl := NewKeyed(10, 3)

	for i := 0; i < 3; i++ {
		if !kl.Allow("user1") {
			t.Fatalf("Allow(user1) = false on call %d, expected true", i+1)
		}
	}

	if kl.Allow("user1") {
		t.Fatal("Allow(user1) should fail after burst exhausted")
	}
}

func TestKeyedLimiter_Wait(t *testing.T) {
	kl := NewKeyed(100, 1)

	// Drain the token for this key.
	kl.Allow("key")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := kl.Wait(ctx, "key"); err != nil {
		t.Fatalf("Wait() error = %v, expected nil", err)
	}
}

func TestKeyedLimiter_Remove(t *testing.T) {
	kl := NewKeyed(10, 1)

	kl.Allow("x")
	if kl.Allow("x") {
		t.Fatal("Allow(x) should fail after burst exhausted")
	}

	kl.Remove("x")

	// After removal, a new limiter is created with full burst.
	if !kl.Allow("x") {
		t.Fatal("Allow(x) should succeed after Remove (fresh limiter)")
	}
}

func TestKeyedLimiter_Size(t *testing.T) {
	kl := NewKeyed(10, 1)

	if kl.Size() != 0 {
		t.Fatalf("Size() = %d, expected 0", kl.Size())
	}

	kl.Allow("a")
	kl.Allow("b")
	kl.Allow("c")

	if kl.Size() != 3 {
		t.Fatalf("Size() = %d, expected 3", kl.Size())
	}

	kl.Remove("b")

	if kl.Size() != 2 {
		t.Fatalf("Size() = %d, expected 2", kl.Size())
	}
}

func TestKeyedLimiter_OnReject(t *testing.T) {
	kl := NewKeyed(10, 1)

	var rejectedKeys []string
	kl.OnReject(func(key string) {
		rejectedKeys = append(rejectedKeys, key)
	})

	// First call succeeds, second is rejected.
	kl.Allow("user1")
	kl.Allow("user1") // rejected

	kl.Allow("user2")
	kl.Allow("user2") // rejected
	kl.Allow("user2") // rejected

	if len(rejectedKeys) != 3 {
		t.Fatalf("OnReject called %d times, expected 3", len(rejectedKeys))
	}
	if rejectedKeys[0] != "user1" {
		t.Fatalf("rejectedKeys[0] = %q, expected %q", rejectedKeys[0], "user1")
	}
	if rejectedKeys[1] != "user2" {
		t.Fatalf("rejectedKeys[1] = %q, expected %q", rejectedKeys[1], "user2")
	}
	if rejectedKeys[2] != "user2" {
		t.Fatalf("rejectedKeys[2] = %q, expected %q", rejectedKeys[2], "user2")
	}
}

func TestKeyedLimiter_OnReject_Nil(t *testing.T) {
	kl := NewKeyed(10, 1)

	// Should not panic when no callback is set.
	kl.Allow("a")
	kl.Allow("a") // rejected, no callback

	// Set and then clear.
	called := false
	kl.OnReject(func(key string) { called = true })
	kl.OnReject(nil)

	kl.Allow("b")
	kl.Allow("b") // rejected, callback cleared

	if called {
		t.Fatal("OnReject callback was called after being cleared")
	}
}

func TestKeyedLimiter_RemoveNonExistentKey(t *testing.T) {
	kl := NewKeyed(10, 1)

	// Removing a key that doesn't exist should not panic.
	kl.Remove("nonexistent")

	if kl.Size() != 0 {
		t.Fatalf("Size() = %d, expected 0", kl.Size())
	}
}
