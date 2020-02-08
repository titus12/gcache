package cache

const (
	TYPE_LRU = "LRU"
	TYPE_LFU = "LFU" //TODO::未实现
	TYPE_ARC = "ARC" //TODO::未实现
)

type RemoveReason uint32

type EvictCallback func(key interface{}, value interface{}, reason RemoveReason)

const (
	// Expired means the key is past its LifeWindow.
	// @TODO: Go defaults to 0 so in case we want to return EntryStatus back to the caller Expired cannot be differentiated
	Expired RemoveReason = iota
	// NoSpace means the key is the oldest and the cache size was at its maximum when Set was called, or the
	// entry exceeded the maximum shard size.
	NoSpace
	// Deleted means Delete was called and this key was removed as a result.
	Deleted
	// Clears all
	Clear
)

type entry struct {
	key       interface{}
	value     interface{}
	timestamp uint64
}

type ICache interface {
	// Adds a value to the cache, returns true if an eviction occurred and
	// updates the "recently used"-ness of the key.
	Add(key, value interface{}) bool

	// Returns key's value from the cache and
	// updates the "recently used"-ness of the key. #value, isFound
	Get(key interface{}) (value interface{}, ok bool)

	// Checks if a key exists in cache without updating the recent-ness.
	Contains(key interface{}) (ok bool)

	// Returns key's value without updating the "recently used"-ness of the key.
	Peek(key interface{}) (value interface{}, ok bool)

	// Removes a key from the cache.
	Remove(key interface{}) bool

	// Removes the oldest entry from cache.
	RemoveOldest()

	// Returns a slice of the keys in the cache, from oldest to newest.
	Keys() []interface{}

	// Returns the number of items in the cache.
	Len() int

	// Clears all cache entries.
	Clear()

	// clean up items expired
	CleanUp(currentTimestamp uint64)

	// Resizes cache, returning number evicted
	// Resize(int) int
}
