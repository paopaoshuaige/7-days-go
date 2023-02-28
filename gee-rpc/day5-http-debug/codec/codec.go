package codec

import "io"

// Header 客户端和服务端响应与接收的请求头
type Header struct {
	ServiceMethod string // 服务名和方法名
	Seq           uint64 // 请求的序号，某个请求的id，区分不同请求。
	Error         string
}

// Codec 对消息体进行编解码的接口
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Writer(*Header, interface{}) error
}

// NewCodecFunc Codec构造函数
type NewCodecFunc func(closer io.ReadWriteCloser) Codec

const (
	GobType string = "application/gob"
	// JsonType string = "application/json"
)

// NewCodecFuncMap 通过 Codec 的 string 得到构造函数
var NewCodecFuncMap map[string]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[string]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
