package pokecache

import (
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	cache := NewCache(5 * time.Second)

	key := "test-key"
	val := []byte("test-value")

	cache.Add(key, val)

	cachedVal, found := cache.Get(key)
	if !found {
		t.Fatalf("expected to find key %s", key)
	}
	if string(cachedVal) != string(val) {
		t.Fatalf("expected value %s, got %s", val, cachedVal)
	}
}

func TestReapLoop(t *testing.T) {
	cache := NewCache(5 * time.Millisecond)

	key := "test-key"
	val := []byte("test-value")

	cache.Add(key, val)
	time.Sleep(10 * time.Millisecond)

	_, found := cache.Get(key)
	if found {
		t.Fatalf("expected key %s to be reaped", key)
	}
}
