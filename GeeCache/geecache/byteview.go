package geecache

type ByteView struct {
	b []byte // 存储真实的缓存值
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	// 返回的是拷贝, 防止只读数据被修改
	return v.cloneBytes(v.b)
}

func (v ByteView) cloneBytes(b []byte) []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
