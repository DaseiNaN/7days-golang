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
	Writer     http.ResponseWriter
	Req        *http.Request
	Path       string            // 请求路径
	Method     string            // 请求方法
	Params     map[string]string // 请求参数
	StatusCode int               // 状态码
	handlers   []HandlerFunc     // 中间件
	index      int               // 执行到第几个中间件
	engine     *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
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

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

// 用于构造 html 响应的方法
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	// 根据模板文件名选择模板进行渲染
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
