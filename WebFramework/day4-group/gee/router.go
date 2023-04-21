package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node       // key = Get or Post
	handlers map[string]HandlerFunc // key = GET-xx or Post-xx
}

func parsePattern(pattern string) []string {
	elems := strings.Split(pattern, "/")
	parts := make([]string, 0)

	for _, elem := range elems {
		if elem != "" {
			parts = append(parts, elem)
			if elem[0] == '*' {
				// '*' 模糊匹配直接跳出即可, 后面的不需要继续匹配了
				break
			}
		}
	}
	return parts
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)

	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

// TODO: 没看懂
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path) // 解析要匹配的 parts
	params := make(map[string]string) // 记录 URL 参数

	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0) // 查找 path 对应的叶子结点

	if n != nil {
		parts := parsePattern(n.pattern) // 解析查找到的叶子结点的 parts
		for index, part := range parts {
			if part[0] == ':' {
				// 说明当前段是参数段, 加入到参数 map 中.
				// Current: /p/:lang/doc
				// Search: /p/go/docs
				// {lang: "go"}
				params[part[1:]] = searchParts[index]
			}

			if part[0] == '*' && len(part) > 1 {
				// 说明当前段是模糊匹配段
				// Current: /static/*filepath
				// Search: /static/css/geektutu.css
				// {filepath: "css/geektutu.css"}
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	// 获得路由和参数 map
	n, params := r.getRoute(c.Method, c.Path)

	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
