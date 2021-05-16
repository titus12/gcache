package gcache

import (
	"bitbucket.org/funplus/gcache/cache"
	"bitbucket.org/funplus/gcache/cache/LRU"
	"fmt"
	"time"
)

const (
	// For use with functions that take an expiration time (default not expire).
	NoExpiration time.Duration = 0
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 5 * time.Minute
)

const default_evict_strategy = LRU.Name

type GCache struct {
	name      string
	shards    []*cacheShard
	cc        *Options
	shardMask uint64
	close     chan struct{}
}

func (g *GCache) evictCallback(key interface{}, value interface{}, reason cache.RemoveReason) {
	l.Debugf("cache %s: key %v is evicted, value %v, reason %v", g.name, key, value, reason)
	if g.cc.OnRemoveCallbackFunc != nil {
		go func() {
			defer PrintPanicStack()
			g.cc.OnRemoveCallbackFunc(key, value, reason)
		}()
	}
}

func NewGCache(name string, opts ...Option) (*GCache, error) {
	gcache := &GCache{name: name}
	gcache.cc = NewOptions(opts...)
	gcache.shards = make([]*cacheShard, gcache.cc.Shards)
	gcache.shardMask = uint64(gcache.cc.Shards - 1)
	gcache.close = make(chan struct{})
	setLogger(gcache.cc.Logger)
	if !isPowerOfTwo(int(gcache.cc.Shards)) {
		return nil, fmt.Errorf("Shards number: %d must be power of two", gcache.cc.Shards)
	}

	for i := 0; i < int(gcache.cc.Shards); i++ {
		shard, err := initNewShard(gcache)
		if err != nil {
			return nil, err
		}
		gcache.shards[i] = shard
	}

	if gcache.cc.CleanInterval > 0 && gcache.cc.Expiration > 0 {
		go func() {
			defer PrintPanicStack()
			ticker := time.NewTicker(gcache.cc.CleanInterval)
			defer ticker.Stop()
			for {
				select {
				case t := <-ticker.C:
					gcache.cleanUp(t.Unix())
				case <-gcache.close:
					return
				}
			}
		}()
	}

	return gcache, nil
}

func (c *GCache) getShard(key interface{}) (shard *cacheShard) {
	hashedKey := c.cc.Hasher.Sum64(key)
	return c.shards[hashedKey&c.shardMask]
}

func (c *GCache) Set(key interface{}, entity interface{}) bool {
	shard := c.getShard(key)
	return shard.set(key, entity)
}

// Get reads entry for the key.
// It returns an ErrEntryNotFound when
// no entry exists for the given key.
func (c *GCache) Get(key interface{}) (interface{}, bool) {
	shard := c.getShard(key)
	return shard.get(key)
}

func (c *GCache) Count() int {
	count := 0
	for _, shard := range c.shards {
		count += shard.count()
	}
	return count
}

func (c *GCache) LoadOrStore(key interface{}, entity interface{}) (interface{}, bool) {
	shard := c.getShard(key)
	return shard.loadOrStore(key, entity)
}

func (c *GCache) CompareAndSet(key interface{}, expect, update interface{}, equal func(old, new interface{}) bool) (interface{}, bool) {
	shard := c.getShard(key)
	return shard.compareAndSet(key, expect, update, equal)
}

// Delete removes the key
func (c *GCache) Delete(key interface{}) bool {
	shard := c.getShard(key)
	return shard.remove(key)
}

//Contains contains the key
func (c *GCache) Contains(key interface{}) bool {
	shard := c.getShard(key)
	return shard.contains(key)
}

// clean up keys expired
func (c *GCache) cleanUp(currentTimestamp int64) {
	for _, shard := range c.shards {
		shard.cleanUp(currentTimestamp)
	}
}

// Close is used to signal a shutdown of the cache when you are done with it.
// This allows the cleaning goroutines to exit and ensures references are not
// kept to the cache preventing GC of the entire cache.
func (c *GCache) Close() error {
	close(c.close)
	return nil
}
