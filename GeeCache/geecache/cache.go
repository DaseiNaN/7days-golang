package geecache

import (
	"geecache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex // 互斥锁
	lru        *lru.Cache // LRU 缓存
	cacheBytes int64      // 缓存大小
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		// Lazy Initialization
		// 提高性能并减少内存需求
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	} else {
		if v, ok := c.lru.Get(key); ok {
			return v.(ByteView), ok
		}
	}
	return
}
