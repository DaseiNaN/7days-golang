// Package codec
// 描述: 消息的编码解码器
package codec

import "io"

type Header struct {
	ServiceMethod string // "Service.Method"
	Seq           uint64 // 请求的序号, 用来区分不同的请求; 由客户端选定;
	Error         string // 错误信息, 客户端处为 nil, 服务端将错误放在这里;
}

type Codec interface {
	io.Closer                         // 关闭连接
	ReadHeader(*Header) error         // 读 Header
	ReadBody(interface{}) error       // 读 Body
	Write(*Header, interface{}) error // 写回复
}

type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
