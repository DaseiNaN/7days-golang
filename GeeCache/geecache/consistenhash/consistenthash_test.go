package consistenthash

import (
	"strconv"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func initHashMap() *Map {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("6", "4", "2")
	return hash
}
func TestHashing(t *testing.T) {
	hash := initHashMap()

	c.Convey("HashMap 测试", t, func() {
		tt := []struct {
			name    string
			pattern string
			expect  string
		}{
			{name: "用例 1", pattern: "2", expect: "2"},
			{name: "用例 2", pattern: "11", expect: "2"},
			{name: "用例 3", pattern: "23", expect: "4"},
			{name: "用例 4", pattern: "27", expect: "2"},
		}
		for _, tc := range tt {
			c.Convey(tc.name, func() {
				v := hash.Get(tc.pattern)
				c.So(v, c.ShouldEqual, tc.expect)
			})
		}
	})
}
