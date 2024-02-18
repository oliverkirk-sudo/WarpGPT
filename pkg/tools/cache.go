package tools

import (
	"WarpGPT/pkg/logger"
	"sync"
)

var AllCache Cache

type CacheItem interface{}

type Cache struct {
	items map[string]CacheItem
	lock  sync.Mutex
}

func init() {
	AllCache = Cache{items: make(map[string]CacheItem)}
}
func (c *Cache) CacheSet(key string, value CacheItem) {
	c.lock.Lock()
	logger.Log.Debug("CacheSet")
	defer c.lock.Unlock()

	c.items[key] = value
}

func (c *Cache) CacheGet(key string) (CacheItem, bool) {
	c.lock.Lock()
	logger.Log.Debug("CacheGet")
	defer c.lock.Unlock()
	item, exists := c.items[key]
	return item, exists
}
