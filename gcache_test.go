package GCache

import (
	"GCache/cache"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

type Entry struct {
	name string
	id   int
}

func Test_newGCache(t *testing.T) {
	//c := DefaultConfig(2 * time.Second)
	f := func(key interface{}, value interface{}, reason cache.RemoveReason) {
		fmt.Printf("evict key:%v,reason %v \n", key, reason)
	}
	c := Config{
		Shards:            1,
		defaultExpiration: 200 * time.Second,
		CleanInterval:     300 * time.Second,
		MaxEntrySize:      0,
		EvictType:         cache.TYPE_LRU,
		Hasher:            newDefaultHasher(),
		OnRemoveFunc:      f,
		//Logger:            DefaultLogger(),
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	gcache, _ := NewGCache(c)
	start := time.Now()
	defer func() {
		fmt.Printf("cost: %v", time.Since(start))
	}()
	testEntry := &Entry{"1", 1}
	testEntry2 := &Entry{"2", 2}
	gcache.Set("aaa", testEntry)
	gcache.Set("bbb", testEntry2)
	isContains := gcache.Contains("aaa")
	fmt.Printf("contains v:%v \n", isContains)
	//time.Sleep(4 * time.Second)
	test, _ := gcache.Get("aaa")
	fmt.Printf("get v:%v \n", test)
	test1, _ := gcache.Get("bbb")
	fmt.Printf("get v:%v \n", test1)
	isDeleted := gcache.Delete("aaa")
	fmt.Printf("del v:%v \n", isDeleted)
	isDeleted2 := gcache.Delete("bbb")
	fmt.Printf("del v:%v \n", isDeleted2)
}
