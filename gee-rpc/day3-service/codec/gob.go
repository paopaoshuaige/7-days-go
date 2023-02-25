package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

// GobCodec 消息结构体
type GobCodec struct {
	conn io.ReadWriteCloser // 构建函数传入，建立 socket 获取到的链接实例
	buf  *bufio.Writer      // 防止阻塞，带缓冲的 Writer，可以提升性能
	dec  *gob.Decoder       // 对应 Decoder，给值编码
	enc  *gob.Encoder       // 对应 Encoder，给值解码
}

// 强制类型转换，保证 GobCodec 实现了 Codec
var _ Codec = (*GobCodec)(nil)

// NewGobCodec 初始化
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	// 新建一个写入 io 的带缓冲的 Writer
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

// ReadHeader 返回解码之后的 Header
func (g *GobCodec) ReadHeader(h *Header) error {
	return g.dec.Decode(h)
}

// ReadBody 返回解码之后的 Body
func (g *GobCodec) ReadBody(i interface{}) error {
	return g.dec.Decode(i)
}

// Writer 写入序列化之后的 Header 和 Body
func (g *GobCodec) Writer(h *Header, i interface{}) (err error) {
	defer func() {
		// 把 buf 里面的东西全写入 conn
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err = g.enc.Encode(h); err != nil {
		log.Println("rpc: gob error encoding header:", err)
		return
	}
	if err = g.enc.Encode(i); err != nil {
		log.Println("rpc: gob error encoding body:", err)
		return
	}
	return
}

// Close 关闭 io 链接
func (g *GobCodec) Close() error {
	return g.conn.Close()
}
