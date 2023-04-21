package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 使得构造 Json 数据更加简洁
// interface{} 说明 value 可以为任意类型
type H map[string]interface{}

type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request
	// 请求相关信息
	Path   string
	Method string
	// 请求参数
	Params map[string]string
	// 状态码
	StatusCode int
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

// 访问解析到的参数
func (c *Context) Param(key string) string {
	v, _ := c.Params[key]
	return v
}

// 用于访问表格数据的方法
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// 用于访问 URL 路径参数的方法
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// 用于构造 string 响应的方法
// text/plain
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// 用于构造 json 响应的方法
// application/json
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// 用于构造 data 响应的方法
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// 用于构造 html 响应的方法
// text/html
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
