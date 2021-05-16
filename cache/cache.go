package cache

//warn:这是一个有些危险的操作，上层加锁删除元素，在evicte的回调中如果再次操作cache会陷入死锁
type EvictCallback func(key interface{}, value interface{}, reason RemoveReason)

type EVICT_STRATEGY = string

//evict_strategy_lfu EVICT_STRATEGY = "LFU" //TODO::未实现
//evict_strategy_arc EVICT_STRATEGY = "ARC" //TODO::未实现

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
	CleanUp(currentTimestamp int64)

	// Resizes cache, returning number evicted
	// Resize(int) int
}
