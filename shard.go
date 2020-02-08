package GCache

import (
	"GCache/cache"
	"sync"
)

type cacheShard struct {
	cache      cache.ICache
	lock       sync.RWMutex
	onRemove   cache.EvictCallback
	logger     Logger
	clock      clock
	expiration uint64
}

func initNewShard(config Config, clock clock) *cacheShard {
	var c cache.ICache
	size := config.initialShardSize()
	switch config.EvictType {
	//case TYPE_SIMPLE:
	//newSimpleCache(cb)
	case cache.TYPE_LRU:
		c = cache.NewLRUCache(size, config.defaultExpiration, config.OnRemoveFunc)
	case cache.TYPE_LFU:
		//newLFUCache(cb)
		fallthrough
	case cache.TYPE_ARC:
		//newARC(cb)
		fallthrough
	default:
		panic("gcache: Unknown type " + config.EvictType)
	}

	shard := &cacheShard{
		cache:      c,
		onRemove:   config.OnRemoveFunc,
		logger:     newLogger(config.Logger),
		clock:      clock,
		expiration: uint64(config.defaultExpiration.Seconds()),
	}

	return shard
}

// Get looks up a key's value from the cache.
func (s *cacheShard) get(key string) (value interface{}, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	value, ok = s.cache.Get(key)
	return
}

// Peek returns key's value without updating the "recently used and timestamp"-ness of the key.
func (s *cacheShard) peek(key string) (value interface{}, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	value, ok = s.cache.Peek(key)
	return
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (s *cacheShard) set(key, value interface{}) (evicted bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	evicted = s.cache.Add(key, value)
	return evicted
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

// RemoveOldest removes the oldest item from the cache.
func (s *cacheShard) removeOldest() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache.RemoveOldest()
}

func (s *cacheShard) cleanUp(currentTimestamp uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache.CleanUp(currentTimestamp)
}
