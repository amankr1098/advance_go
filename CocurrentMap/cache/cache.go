package cache

import "sync"

var cacheMutex sync.RWMutex

type Cache[K comparable, V any] struct {
	Data map[K]V
}

func NewCache[K comparable, V any]() Cache[K, V] {
	return Cache[K, V]{
		Data: make(map[K]V),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	c.Data[key] = value
}

func (c *Cache[K, V]) Get(key K) V {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return c.Data[key]
}

func (c *Cache[K, V]) Delete(key K) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	delete(c.Data, key)
}

func (c *Cache[K, V]) GetKeys() []K {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	keys := make([]K, 0, len(c.Data))
	for k := range c.Data {
		keys = append(keys, k)
	}
	return keys
}
