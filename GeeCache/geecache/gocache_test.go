package geecache

import (
	"fmt"
	"log"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestGetter(t *testing.T) {
	var f Getter
	f = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	c.Convey("Getter 测试", t, func() {
		v, _ := f.Get("key")
		c.So(v, c.ShouldResemble, expect)
	})
}

func initTestGroup() (*Group, *map[string]int) {
	var db map[string]string = make(map[string]string)
	db["Tom"] = "630"
	db["Jack"] = "589"
	db["Sam"] = "567"

	var loadCounts map[string]int = make(map[string]int, len(db))
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	return gee, &loadCounts
}
func TestGet(t *testing.T) {
	g, lc := initTestGroup()
	c.Convey("Get 测试", t, func() {
		tt := []struct {
			name    string
			pattern string
			expect  interface{}
			hit     int
		}{
			{name: "key 合法但不存在于缓存", pattern: "Tom", expect: "630", hit: 1},
			{name: "key 合法但且存在于缓存", pattern: "Tom", expect: "630", hit: 1},
			{name: "key 不存在", pattern: "Tomas", expect: ""},
		}
		for _, tc := range tt {
			c.Convey(tc.name, func() {
				if v, err := g.Get(tc.pattern); err != nil {
					c.So(v.String(), c.ShouldEqual, tc.expect)
				} else {
					c.So(v.String(), c.ShouldEqual, tc.expect)
					c.So((*lc)[tc.pattern], c.ShouldEqual, tc.hit)
				}
			})
		}
	})
}
