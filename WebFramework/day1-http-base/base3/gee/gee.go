package gee

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

type Engine struct {
	// 映射表
	// key: method-pattern
	// value: handler 函数
	router map[string]HandlerFunc
}

// 构造函数, 返回指针
func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

func (e *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	// method: Get / Post
	// pattern: URL 地址
	// handler: handler 函数
	key := fmt.Sprintf("%s-%s", method, pattern)
	e.router[key] = handler
}

// 用户调用 (*Engine).GET() 方法时, 会将路由和 handler 注册到 Engine 的映射表 router 中.
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRoute("GET", pattern, handler)
}

// 用户调用 (*Engine).POST() 方法时, 会将路由和 handler 注册到 Engine 的映射表 router 中.
func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRoute("POST", pattern, handler)
}

// ListenAndServe 的封装
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// 实现 ServeHTTP 接口, 拦截 HTTP 请求
// 解析请求路径 pattern -> 查找映射表 router -> 执行 / 404
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := fmt.Sprintf("%s-%s", req.Method, req.URL.Path)
	if handler, ok := e.router[key]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}
