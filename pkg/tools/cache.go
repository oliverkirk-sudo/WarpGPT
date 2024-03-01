package tools

import (
	"WarpGPT/pkg/logger"
	"sync"
	"time"
)

type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

type Cache struct {
	items map[string]CacheItem
	lock  sync.Mutex
}

var AllCache Cache

func init() {
	AllCache = Cache{items: make(map[string]CacheItem)}
}

func (c *Cache) CacheSet(key string, value CacheItem, expiration time.Duration) {
    c.lock.Lock()
    defer c.lock.Unlock()
		value.ExpiresAt = time.Now().Add(expiration)
    c.items[key] = value
	logger.Log.Debug("CacheSet: Key =", key, "Expiration =", expiration, "Data =", value.Data)
}

func (c *Cache) CacheGet(key string) (CacheItem, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	item, exists := c.items[key]
	if exists && item.ExpiresAt.After(time.Now()) {
		logger.Log.Debug("CacheGet (Hit): Key =", key, "Expiration =", item.ExpiresAt, "Data =", item.Data)
		return item, true
	}
	logger.Log.Debug("CacheGet (Miss): Key =", key)
	return CacheItem{}, false
}
