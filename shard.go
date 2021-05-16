package gcache

import (
	cache2 "bitbucket.org/funplus/gcache/cache"
	"fmt"
	"sync"
)

type cacheShard struct {
	cache      cache2.ICache
	lock       sync.RWMutex
	expiration uint64
}

const minimumEntriesInShard = 10

func initNewShard(c *GCache) (*cacheShard, error) {
	opts := c.cc
	size := max(opts.MaxEntrySize/uint32(opts.Shards), minimumEntriesInShard)
	cacheBuilder := cache2.Get(opts.EvictStrategy)
	if cacheBuilder == nil {
		return nil, fmt.Errorf("gcache: cache unregistered %s", opts.EvictStrategy)
	}
	cacheImpl := cacheBuilder.Build(size, opts.Expiration, c.evictCallback)
	shard := &cacheShard{
		cache:      cacheImpl,
		expiration: uint64(opts.Expiration.Seconds()),
	}
	return shard, nil
}

// Get looks up a key's value from the cache.
func (s *cacheShard) get(key interface{}) (value interface{}, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	value, ok = s.cache.Get(key)
	return
}

// Peek returns key's value without updating the "recently used and timestamp"-ness of the key.
func (s *cacheShard) peek(key interface{}) (value interface{}, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	value, ok = s.cache.Peek(key)
	return
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (s *cacheShard) set(key, value interface{}) (ok bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	ok = s.cache.Add(key, value)
	return
}

func (s *cacheShard) loadOrStore(key interface{}, newValue interface{}) (value interface{}, ok bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	value, ok = s.cache.Get(key)
	if ok {
		return
	}
	s.cache.Add(key, newValue)
	return
}

func (s *cacheShard) compareAndSet(key interface{}, expect, update interface{}, equal func(old, new interface{}) bool) (interface{}, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, ok := s.cache.Get(key)
	if !ok || equal(v, expect) {
		s.cache.Add(key, update)
		return update, true
	}
	return v, false
}

// Remove removes the provided key from the cache.
func (s *cacheShard) remove(key interface{}) (present bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	present = s.cache.Remove(key)
	return
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (s *cacheShard) contains(key interface{}) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.cache.Contains(key)
}

// Count returns the number of items in the cache.
func (s *cacheShard) count() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.cache.Len()
}

// RemoveOldest removes the oldest item from the cache.
func (s *cacheShard) removeOldest() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache.RemoveOldest()
}

func (s *cacheShard) cleanUp(currentTimestamp int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache.CleanUp(currentTimestamp)
}
