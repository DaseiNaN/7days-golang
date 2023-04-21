package singleflight

import (
	"sync"
)

// 正在进行中的，或者已经结束的请求.
type call struct {
	wg  sync.WaitGroup // 避免重入
	val interface{}
	err error
}

// 管理不同 key 的请求 call.
type Group struct {
	mu sync.Mutex // 保护 m 不会被并发读写
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		// 当前 key 对应的请求正在处理 or 已经处理过了
		g.mu.Unlock()
		c.wg.Wait()         // 有请求正在进行, 等待
		return c.val, c.err // 请求结束, 返回结果
	}

	c := new(call)
	c.wg.Add(1)  // 发起请求前加锁
	g.m[key] = c // 将请求添加到 g.m 中, 表示已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() // 调用 function 发起请求
	c.wg.Done()         // 请求结束释放锁

	g.mu.Lock()
	delete(g.m, key) // 更新 g.m
	g.mu.Unlock()

	return c.val, c.err
}
