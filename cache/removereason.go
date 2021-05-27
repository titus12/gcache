package cache

//go:generate stringer -type=RemoveReason
type RemoveReason uint32

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
