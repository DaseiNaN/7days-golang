package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // 注入式 Hash 函数
	replicas int            // 虚拟节点副本数
	keys     []int          // 哈希环
	hashMap  map[int]string // 虚拟节点和真实节点的映射表, key: 虚拟节点的 Hash 值, value: 真实节点的名称
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	// 添加真实节点
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 计算虚拟节点的 Hash 值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 将虚拟节点加入到 Hash 环中
			m.keys = append(m.keys, hash)
			// 简历虚拟节点和真实节点之间的映射关系
			m.hashMap[hash] = key
		}
	}
	// 对 Hash 环中的值进行排序
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	// 计算要查找的节点的 Hash 值
	hash := int(m.hash([]byte(key)))
	// 找到距离最近的虚拟节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]

}
