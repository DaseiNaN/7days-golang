package gee

import (
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	c.Convey("路径解析测试", t, func() {
		tt := []struct {
			name    string
			pattern string
			expect  []string
		}{
			{"路径中包含 ':'", "/p/:name", []string{"p", ":name"}},
			{"路径中包含 '*'", "/p/*", []string{"p", "*"}},
			{"路径中包含多个 '*'", "/p/*name/*", []string{"p", "*name"}},
		}
		for _, tc := range tt {
			c.Convey(tc.name, func() {
				got := parsePattern(tc.pattern)
				c.So(got, c.ShouldResemble, tc.expect)
			})
		}
	})
}

func TestGetRoot(t *testing.T) {
	r := newTestRouter()
	c.Convey("路由获取测试", t, func() {
		n, params := r.getRoute("GET", "/hello/dasein")
		c.Convey("路由匹配", func() {
			c.So(n, c.ShouldNotBeNil)
		})
		c.Convey("路径获取", func() {
			c.So(n.pattern, c.ShouldEqual, "/hello/:name")
		})
		c.Convey("参数获取", func() {
			c.So(params["name"], c.ShouldEqual, "dasein")
		})
	})

}
