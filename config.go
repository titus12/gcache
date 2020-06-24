package GCache

import (
	"github.com/titus12/gcache/cache"
	"time"
)

const (
	minimumEntriesInShard = 10 // Minimum number of entries in single shard
)

type Config struct {
	// Number of cache shards, value must be a power of two
	Shards int
	// Time after which entry can be evicted
	Expiration time.Duration
	// Interval between removing expired entries (clean up).
	// If set to <= 0 then no action is performed. Setting to < 1 second is counterproductive — gcache has a one second resolution.
	CleanInterval time.Duration
	// Max number of entries in life window. Used only to calculate initial size for cache shards.
	// When proper value is set then additional memory allocation does not occur.
	MaxEntrySize int
	// Type of evict for cache, also its build type.
	EvictType string
	// Hasher used to map between string keys and unsigned 64bit integers, by default fnv64 hashing is used.
	Hasher Hasher
	// OnRemove is a callback fired when the oldest entry is removed because of its expiration time or no space left
	// for the new entry, or because delete was called.
	// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
	// ignored if OnRemoveWithMetadata is specified.
	OnRemoveFunc cache.EvictCallback
	// Logger is a logging interface and used in combination with `Verbose`
	// Defaults to `DefaultLogger()`
	Logger Logger
}

// DefaultConfig initializes config with default values.
// When load for BigCache can be predicted in advance then it is better to use custom config.
func DefaultConfig(eviction time.Duration) Config {
	return Config{
		Shards:        1024,
		Expiration:    eviction,
		EvictType:     cache.TYPE_LRU,
		CleanInterval: 180 * time.Second,
		MaxEntrySize:  1024 * 1024,
		Hasher:        newDefaultHasher(),
		Logger:        DefaultLogger(),
	}
}

func (c Config) initialShardSize() int {
	return max(c.MaxEntrySize/c.Shards, minimumEntriesInShard)
}
