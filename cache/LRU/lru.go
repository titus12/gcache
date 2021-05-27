package LRU

import (
	"bitbucket.org/funplus/gcache/cache"
	"container/list"
	"time"
)

const Name = cache.EVICT_STRATEGY("LRU")

type lRUCacheBuilder struct{}

func NewBuilder() cache.CacheBuilder {
	return &lRUCacheBuilder{}
}

func init() {
	cache.Register(NewBuilder())
}

func (*lRUCacheBuilder) Build(maxEntries uint32, expiration time.Duration, onEvict cache.EvictCallback) cache.ICache {
	return newLRUCache(maxEntries, expiration, onEvict)
}

func (*lRUCacheBuilder) Name() string {
	return Name
}

// Cache is an LRUCache cache. It is not safe for concurrent access.
type LRUCache struct {
	// size is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	size uint32
	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	onEvicted  cache.EvictCallback
	evictList  *list.List
	items      map[interface{}]*list.Element
	expiration time.Duration
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func newLRUCache(maxEntries uint32, expiration time.Duration, onEvict cache.EvictCallback) *LRUCache {
	return &LRUCache{
		size:       maxEntries,
		evictList:  list.New(),
		items:      make(map[interface{}]*list.Element, maxEntries),
		expiration: expiration,
		onEvicted:  onEvict,
	}
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRUCache) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*cache.Entry).Key
		i++
	}
	return keys
}

// Add adds a value to the cache, returns true if an eviction occurred.
func (c *LRUCache) Add(key interface{}, value interface{}) bool {
	if c.items == nil {
		c.items = make(map[interface{}]*list.Element)
		c.evictList = list.New()
	}
	expireAt := time.Now().Add(c.expiration).Unix()
	if ee, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ee)
		kv := ee.Value.(*cache.Entry)
		kv.Value = value
		kv.Timestamp = expireAt
		//atomic.StoreInt64(&kv.timestamp, expireAt)
		return true
	}
	ele := c.evictList.PushFront(&cache.Entry{key, value, expireAt})
	c.items[key] = ele
	if c.size != 0 && uint32(c.evictList.Len()) > c.size {
		c.RemoveOldest()
	}
	return true
}

// Get looks up a key's value from the cache
func (c *LRUCache) Get(key interface{}) (value interface{}, ok bool) {
	if c.items == nil {
		return
	}
	if ele, hit := c.items[key]; hit {
		//expireAt := atomic.LoadInt64(&ele.Value.(*entry).timestamp)
		kv := ele.Value.(*cache.Entry)
		expireAt := ele.Value.(*cache.Entry).Timestamp
		if c.expiration > 0 && time.Now().Unix() > expireAt {
			return kv.Value, false
		}
		c.evictList.MoveToFront(ele)
		return kv.Value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *LRUCache) Remove(key interface{}) bool {
	if c.items == nil {
		return false
	}
	if ele, hit := c.items[key]; hit {
		c.removeElement(ele, cache.Deleted)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRUCache) RemoveOldest() {
	if c.items == nil {
		return
	}
	ele := c.evictList.Back()
	if ele != nil {
		c.removeElement(ele, cache.NoSpace)
	}
}

func (c *LRUCache) removeElement(e *list.Element, reason cache.RemoveReason) {
	c.evictList.Remove(e)
	kv := e.Value.(*cache.Entry)
	delete(c.items, kv.Key)
	c.onEvicted(kv.Key, kv.Value, reason)
}

// Len returns the number of items in the cache.
func (c *LRUCache) Len() int {
	if c.items == nil {
		return 0
	}
	return c.evictList.Len()
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRUCache) Peek(key interface{}) (value interface{}, ok bool) {
	var ele *list.Element
	if ele, ok = c.items[key]; ok {
		//expireAt := atomic.LoadInt64(&ent.Value.(*entry).timestamp)
		kv := ele.Value.(*cache.Entry)
		expireAt := ele.Value.(*cache.Entry).Timestamp
		if c.expiration > 0 && time.Now().Unix() > expireAt {
			return kv.Value, false
		}
		return kv.Value, true
	}
	return nil, ok
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRUCache) Contains(key interface{}) bool {
	ele, ok := c.items[key]
	if ok {
		//expireAt := atomic.LoadInt64(&ele.Value.(*entry).timestamp)
		expireAt := ele.Value.(*cache.Entry).Timestamp
		if c.expiration > 0 && time.Now().Unix() > expireAt {
			return false
		}
	}
	return ok
}

func (c *LRUCache) CleanUp(currentTimestamp int64) {
	if c.expiration == 0 {
		return
	}
	for _, e := range c.items {
		//expireAt := atomic.LoadInt64(&e.Value.(*entry).timestamp)
		expireAt := e.Value.(*cache.Entry).Timestamp
		if currentTimestamp > expireAt {
			c.removeElement(e, cache.Expired)
		}
	}
}

// Clear purges all stored items from the cache.
func (c *LRUCache) Clear() {
	for _, e := range c.items {
		kv := e.Value.(*cache.Entry)
		c.onEvicted(kv.Key, kv.Value, cache.Clear)
	}
	c.evictList = nil
	c.items = nil
}
