package main

import (
	"bytes"
	"encoding/json"
	"example/gee"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func newTestEngine() *gee.Engine {
	r := gee.New()
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})

	r.GET("/hello", func(c *gee.Context) {
		// /hello?name=dasein
		c.String(http.StatusOK, "hello %s, you are at %s\n", c.Query("name"), c.Path)
	})

	r.GET("/hello/:name", func(c *gee.Context) {
		// /hello/dasein
		c.String(http.StatusOK, "hello %s, you are at %s\n", c.Param("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"filepath": c.Param("filepath"),
		})
	})

	r.POST("/login", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})
	return r
}

func TestMain(t *testing.T) {
	r := newTestEngine()
	c.Convey("请求测试", t, func() {
		tt := []struct {
			name    string
			method  string
			pattern string
			expect  interface{}
		}{
			{"?name 参数", "GET", "/hello/?name=dasein", "hello dasein, you are at /hello/\n"},
			{":name 解析", "GET", "/hello/dasein", "hello dasein, you are at /hello/dasein\n"},
			{"*filepath 解析", "GET", "/assets/css/dasein-harbour.js", map[string]string{"filepath": "css/dasein-harbour.js"}},
		}
		for _, tc := range tt {
			c.Convey(tc.name, func() {
				req := httptest.NewRequest(tc.method, tc.pattern, nil)
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)
				if resp.Header().Get("Content-Type") != "application/json" {
					content := new(bytes.Buffer)
					io.Copy(content, resp.Body)
					c.So(content.String(), c.ShouldEqual, tc.expect)
				} else {
					var content map[string]string
					json.NewDecoder(resp.Body).Decode(&content)
					c.So(content, c.ShouldResemble, tc.expect)
				}
			})
		}
	})
}
