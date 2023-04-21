package gee

import (
	"log"
	"net/http"
)

type HandlerFunc func(*Context)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // 中间件
	parent      *RouterGroup  // 嵌套 Group
	engine      *Engine       // 统一的 HTTP Engine
}

type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup // 存储所有的 Group
}

// 改写 New 函数
func New() *Engine {
	e := &Engine{router: newRouter()}
	e.RouterGroup = &RouterGroup{engine: e}
	e.groups = []*RouterGroup{e.RouterGroup}
	return e
}

// 创建一个新的 RouterGroup
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	e := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: e,
	}
	e.groups = append(e.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) Get(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// func (e *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
// 	// method: Get or Post
// 	// pattern: monitor uri
// 	// handler: uri handler function
// 	e.router.addRoute(method, pattern, handler)
// }

// func (e *Engine) GET(pattern string, handler HandlerFunc) {
// 	e.addRoute("GET", pattern, handler)
// }

// func (e *Engine) POST(pattern string, handler HandlerFunc) {
// 	e.addRoute("POST", pattern, handler)
// }

func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// ServeHTTP 首先将请求封装称 Context, 简化接口调用
	c := newContext(w, req)
	e.router.handle(c)
}
