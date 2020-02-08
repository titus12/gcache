package cache

import (
	"container/list"
	"sync/atomic"
	"time"
)

// Cache is an LRUCache cache. It is not safe for concurrent access.
type LRUCache struct {
	// size is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	size int
	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	onEvicted  EvictCallback
	ll         *list.List
	items      map[interface{}]*list.Element
	expiration uint64
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func NewLRUCache(maxEntries int, eviction time.Duration, onEvict EvictCallback) *LRUCache {
	return &LRUCache{
		size:       maxEntries,
		ll:         list.New(),
		items:      make(map[interface{}]*list.Element, maxEntries),
		expiration: uint64(eviction.Seconds()),
		onEvicted:  onEvict,
	}
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRUCache) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.ll.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

// Add adds a value to the cache.
func (c *LRUCache) Add(key interface{}, value interface{}) bool {
	if c.items == nil {
		c.items = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	now := uint64(time.Now().Unix())
	if ee, ok := c.items[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		ee.Value.(*entry).timestamp = now
		return true
	}
	ele := c.ll.PushFront(&entry{key, value, now})
	c.items[key] = ele
	if c.size != 0 && c.ll.Len() > c.size {
		c.RemoveOldest()
	}
	return true
}

// Get looks up a key's value from the cache,update timestamp
func (c *LRUCache) Get(key interface{}) (value interface{}, ok bool) {
	if c.items == nil {
		return
	}
	if ele, hit := c.items[key]; hit {
		now := uint64(time.Now().Unix())
		if c.expiration > 0 && now-ele.Value.(*entry).timestamp > c.expiration {
			return nil, false
		}
		kv := ele.Value.(*entry)
		atomic.StoreUint64(&kv.timestamp, now)
		c.ll.MoveToFront(ele)
		return kv.value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *LRUCache) Remove(key interface{}) bool {
	if c.items == nil {
		return false
	}
	if ele, hit := c.items[key]; hit {
		c.removeElement(ele, Deleted)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRUCache) RemoveOldest() {
	if c.items == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele, NoSpace)
	}
}

func (c *LRUCache) removeElement(e *list.Element, reason RemoveReason) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.items, kv.key)
	if c.onEvicted != nil {
		c.onEvicted(kv.key, kv.value, reason)
	}
}

// Len returns the number of items in the cache.
func (c *LRUCache) Len() int {
	if c.items == nil {
		return 0
	}
	return c.ll.Len()
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRUCache) Peek(key interface{}) (value interface{}, ok bool) {
	var ent *list.Element
	if ent, ok = c.items[key]; ok {
		now := uint64(time.Now().Unix())
		if c.expiration > 0 && now-ent.Value.(*entry).timestamp > c.expiration {
			return nil, false
		}
		return ent.Value.(*entry).value, true
	}
	return nil, ok
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRUCache) Contains(key interface{}) bool {
	ele, ok := c.items[key]
	if ok {
		now := uint64(time.Now().Unix())
		if c.expiration > 0 && now-ele.Value.(*entry).timestamp > c.expiration {
			return false
		}
	}
	return ok
}

func (c *LRUCache) CleanUp(currentTimestamp uint64) {
	if c.expiration == 0 {
		return
	}
	for _, e := range c.items {
		if currentTimestamp-e.Value.(*entry).timestamp > c.expiration {
			c.removeElement(e, Expired)
		}
	}
}

// Clear purges all stored items from the cache.
func (c *LRUCache) Clear() {
	if c.onEvicted != nil {
		for _, e := range c.items {
			kv := e.Value.(*entry)
			c.onEvicted(kv.key, kv.value, Clear)
		}
	}
	c.ll = nil
	c.items = nil
}
