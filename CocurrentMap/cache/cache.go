package cache

import "sync"

// var

type Cache[K comparable, V any] struct {
	cacheMutex sync.RWMutex
	Data       map[K]V
}

func NewCache[K comparable, V any]() Cache[K, V] {
	return Cache[K, V]{
		cacheMutex: sync.RWMutex{},
		Data:       make(map[K]V),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.Data[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()
	d, ok := c.Data[key]
	return d, ok
}

func (c *Cache[K, V]) Delete(key K) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	delete(c.Data, key)
}

func (c *Cache[K, V]) GetKeys() []K {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()
	keys := make([]K, 0, len(c.Data))
	for k := range c.Data {
		keys = append(keys, k)
	}
	return keys
}
