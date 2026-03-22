package cache

import (
	"sync"
	"testing"
)

func TestCacheBasicOperations(t *testing.T) {
	cache := NewCache[string, int]()

	// Test Set and Get
	cache.Set("foo", 42)
	if val := cache.Get("foo"); val != 42 {
		t.Errorf("expected 42, got %v", val)
	}

	// Test Overwrite
	cache.Set("foo", 100)
	if val := cache.Get("foo"); val != 100 {
		t.Errorf("expected 100, got %v", val)
	}

	// Test Delete
	cache.Delete("foo")
	if val := cache.Get("foo"); val != 0 {
		t.Errorf("expected 0 after delete, got %v", val)
	}

	// Test Keys
	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)
	keys := cache.GetKeys()
	keySet := make(map[string]struct{})
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	if len(keySet) != 3 {
		t.Errorf("expected 3 keys, got %v", keys)
	}
	for _, k := range []string{"a", "b", "c"} {
		if _, ok := keySet[k]; !ok {
			t.Errorf("expected key %s in keys, got %v", k, keys)
		}
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	cache := NewCache[int, int]()
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.Set(i, i*i)
		}(i)
	}
	wg.Wait()
	for i := 0; i < 100; i++ {
		if val := cache.Get(i); val != i*i {
			t.Errorf("expected %d, got %d", i*i, val)
		}
	}
}
