package gcache

import (
	"bitbucket.org/funplus/gcache/cache"
	"time"
)

//go:generate optionGen --option_with_struct_name=false --v=true
func OptionsOptionDeclareWithDefault() interface{} {
	return map[string]interface{}{
		// Number of cache Shards, value must be a power of two
		"Shards": int32(1024),
		// Time after which entry can be evicted
		"Expiration": time.Duration(DefaultExpiration),
		// Type of evict for cache, also its build type.
		"EvictStrategy": cache.EVICT_STRATEGY(default_evict_strategy),
		// Interval between removing expired entries (clean up).
		// If set to <= 0 then no action is performed. Setting to < 1 second is counterproductive — gcache has a one second resolution.
		"CleanInterval": time.Duration(30 * time.Second),
		// Max number of entries in life window. Used only to calculate initial size for cache Shards.
		// When proper value is set then additional memory allocation does not occur.
		"MaxEntrySize": uint32(1024 * 1024),
		// Hasher used to map between string keys and unsigned 64bit integers, by default fnv64 hashing is used.
		"Hasher": (Hasher)(newDefaultHasher()),
		// OnRemove is a callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		// ignored if OnRemoveWithMetadata is specified.
		"OnRemoveCallbackFunc": (cache.EvictCallback)(nil),
		// Development output evicted logs in development mode
		"Development": bool(true),
		// Logger is a logging interface and used in combination with `Verbose`
		// Defaults to `DefaultLogger()`
		"Logger": (Logger)(nil),
	}
}
