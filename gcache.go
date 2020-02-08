package GCache

import (
	"fmt"
	"time"
)

const (
	// For use with functions that take an expiration time (default not expire).
	NoExpiration time.Duration = 0
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 5 * 60
)

type GCache struct {
	shards       []*cacheShard
	clock        clock
	hash         Hasher
	config       Config
	shardMask    uint64
	maxShardSize uint32
	close        chan struct{}
}

func NewGCache(config Config) (*GCache, error) {
	return newGCache(config, &systemClock{})
}

func newGCache(config Config, clock clock) (*GCache, error) {
	if !isPowerOfTwo(config.Shards) {
		return nil, fmt.Errorf("Shards number must be power of two")
	}

	if config.Hasher == nil {
		config.Hasher = newDefaultHasher()
	}

	gcache := &GCache{
		shards:    make([]*cacheShard, config.Shards),
		clock:     clock,
		hash:      config.Hasher,
		config:    config,
		shardMask: uint64(config.Shards - 1),
		close:     make(chan struct{}),
	}

	for i := 0; i < config.Shards; i++ {
		gcache.shards[i] = initNewShard(config, clock)
	}

	if config.CleanInterval > 0 {
		go func() {
			ticker := time.NewTicker(config.CleanInterval)
			defer ticker.Stop()
			for {
				select {
				case t := <-ticker.C:
					gcache.cleanUp(uint64(t.Unix()))
				case <-gcache.close:
					return
				}
			}
		}()
	}

	return gcache, nil
}

func (c *GCache) Set(key string, entity interface{}) bool {
	hashedKey := c.hash.Sum64(key)
	shard := c.getShard(hashedKey)
	return shard.set(key, entity)
}

/*func (c *GCache) SetWithTTL(key string, entity interface{},) bool {

}*/

func (c *GCache) getShard(hashedKey uint64) (shard *cacheShard) {
	return c.shards[hashedKey&c.shardMask]
}

// Get reads entry for the key.
// It returns an ErrEntryNotFound when
// no entry exists for the given key.
func (c *GCache) Get(key string) (interface{}, bool) {
	hashedKey := c.hash.Sum64(key)
	shard := c.getShard(hashedKey)
	return shard.get(key)
}

// Delete removes the key
func (c *GCache) Delete(key string) bool {
	hashedKey := c.hash.Sum64(key)
	shard := c.getShard(hashedKey)
	return shard.remove(key)
}

//Contains contains the key
func (c *GCache) Contains(key string) bool {
	hashedKey := c.hash.Sum64(key)
	shard := c.getShard(hashedKey)
	return shard.contains(key)
}

// clean up keys expired
func (c *GCache) cleanUp(currentTimestamp uint64) {
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
