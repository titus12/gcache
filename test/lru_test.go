package test

import (
	"bitbucket.org/funplus/gcache"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type TestEntry struct {
	name string
	id   int
}

func Test_LRUCache(t *testing.T) {
	Convey("new gCache test", t, func() {
		gcache, err := gcache.NewGCache("test")
		So(err, ShouldBeNil)
		start := time.Now()
		defer func() {
			fmt.Printf("cost: %v", time.Since(start))
		}()
		testEntry := &TestEntry{"1", 1}
		testEntry2 := &TestEntry{"2", 2}
		ok := gcache.Set("aaa", testEntry)
		So(ok, ShouldBeTrue)
		ok = gcache.Set("bbb", testEntry2)
		So(ok, ShouldBeTrue)
		ok = gcache.Contains("aaa")
		So(ok, ShouldBeTrue)
		test, ok := gcache.Get("aaa")
		So(ok, ShouldBeTrue)
		fmt.Printf("get v:%v \n", test)
		ok = gcache.Delete("aaa")
		So(ok, ShouldBeTrue)
		ok = gcache.Delete("bbb")
		So(ok, ShouldBeTrue)
	})
}
