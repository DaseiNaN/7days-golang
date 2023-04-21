package lru

import (
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("dasein"))
	c.Convey("Get 测试", t, func() {
		tt := []struct {
			name     string
			pattern  string
			expected interface{}
		}{
			{name: "key 存在且 value 正确", pattern: "key1", expected: "dasein"},
			{name: "key 不存在", pattern: "key2", expected: nil},
		}
		for _, tc := range tt {
			c.Convey(tc.name, func() {
				if v, ok := lru.Get(tc.pattern); ok {
					c.So(string(v.(String)), c.ShouldEqual, tc.expected)
				} else {
					c.So(ok, c.ShouldBeFalse)
				}
			})
		}
	})
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)

	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	c.Convey("删除最近少使用的节点测试", t, func() {
		_, ok := lru.Get(k1)
		c.So(ok, c.ShouldBeFalse)
		c.So(lru.Len(), c.ShouldEqual, 2)
	})

}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	k1, k2, k3, k4 := "k1", "k2", "k3", "k4"
	v1, v2, v3, v4 := "value1", "value2", "value3", "value4"
	cap := len(k1 + k2 + v1 + v2)

	lru := New(int64(cap), callback)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))
	lru.Add(k4, String(v4))

	expect := []string{"k1", "k2"}
	c.Convey("测试回调函数", t, func() {
		c.So(keys, c.ShouldResemble, expect)
	})
}
