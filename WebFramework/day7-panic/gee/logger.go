package gee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		// 中间件前半部分处理逻辑
		t := time.Now()

		// 等待请求处理结束
		c.Next()

		// 中间件后半部分处理逻辑
		log.Printf("[%d] %s in %s", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
