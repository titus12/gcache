package cache

import (
	"strings"
	"time"
)

var (
	// m is a map from name to cache builder.
	m = make(map[string]CacheBuilder)
)

func Register(b CacheBuilder) {
	m[strings.ToLower(b.Name())] = b
}

type CacheBuilder interface {
	Build(maxEntries uint32, expiration time.Duration, onEvict EvictCallback) ICache
	Name() string
}

func Get(name string) CacheBuilder {
	if b, ok := m[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}
