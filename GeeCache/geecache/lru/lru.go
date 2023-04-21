package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         // 最大使用内存
	nbytes    int64                         // 当前使用内存
	ll        *list.List                    // 双向链表
	cache     map[string]*list.Element      // key: 字符串; *list.Element 指向双向列表节点的指针
	OnEvicted func(key string, value Value) // 删除某个记录时的回调函数
}

type entry struct { // 双向列表节点的数据类型
	key   string
	value Value
}

type Value interface {
	Len() int // 计算占用内存大小
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 这里约定 Front 为队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	// 这里约定 Back 为队首
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		// list 存储的是任意类型, interface类型转换是.(被转换类型)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += -int64(kv.value.Len()) + int64(value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// 查看缓存内有多少数据
func (c *Cache) Len() int {
	return c.ll.Len()
}
