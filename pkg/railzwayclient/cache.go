package railzwayclient

import (
	"sync"
	"time"
)

type cacheItem struct {
	value     any
	expiresAt time.Time
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
	ttl   time.Duration
	limit int
}

func NewCache(limit int, ttl time.Duration) *Cache {
	if limit <= 0 || ttl <= 0 {
		return &Cache{items: nil}
	}

	return &Cache{
		items: make(map[string]cacheItem, limit),
		ttl:   ttl,
		limit: limit,
	}
}

func (c *Cache) Get(key string) (any, bool) {
	if c.items == nil {
		return nil, false
	}

	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(item.expiresAt) {
		if ok {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
		}
		return nil, false
	}

	return item.value, true
}

func (c *Cache) Set(key string, value any) {
	if c.items == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// simple eviction: if over limit, reset cache
	if len(c.items) >= c.limit {
		c.items = make(map[string]cacheItem, c.limit)
	}

	c.items[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}
