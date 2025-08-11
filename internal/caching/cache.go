package cache

import (
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
}

type Cache struct {
	data      sync.Map
	group     singleflight.Group
	itemCount int32
}

func (c *Cache) GetOrCreate(key string, createFn func(entry *CacheEntry) (interface{}, error)) (interface{}, error) {
	if value, ok := c.data.Load(key); ok {
		cacheEntry := value.(CacheEntry)
		if cacheEntry.Expiration.After(time.Now()) {
			return cacheEntry.Value, nil
		} else {
			c.CleanUp()
		}
	}

	value, err, _ := c.group.Do(key, func() (interface{}, error) {
		if value, ok := c.data.Load(key); ok {
			cacheEntry := value.(CacheEntry)
			if cacheEntry.Expiration.After(time.Now()) {
				return cacheEntry.Value, nil
			}
		}

		entry := &CacheEntry{
			Expiration: time.Now().Add(time.Hour),
		}

		v, err := createFn(entry)
		if err != nil {
			return nil, err
		}

		entry.Value = v
		c.data.Store(key, *entry)
		atomic.AddInt32(&c.itemCount, 1)
		return v, nil
	})

	return value, err
}

func (c *Cache) CleanUp() {
	if atomic.LoadInt32(&c.itemCount) == 0 {
		return
	}

	c.data.Range(func(key, value interface{}) bool {
		entry := value.(CacheEntry)
		if entry.Expiration.Before(time.Now()) {
			c.data.Delete(key)
			atomic.AddInt32(&c.itemCount, -1)
		}
		return true
	})
}
