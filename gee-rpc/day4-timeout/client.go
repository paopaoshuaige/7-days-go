package geerpc

import (
	"day3-service/codec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// Call rpc调用所需要的信息
type Call struct {
	Seq           uint64      // 序列号
	ServiceMethod string      // 格式 "<service>.<method>"
	Args          interface{} // 函数的参数
	Reply         interface{} // 函数的回复
	Error         error
	Done          chan *Call // 调用结束用来通知调用方
}

// done 把c发送到 Done 里
func (c *Call) done() {
	c.Done <- c
}

// Client 客户端字段，可能有多个未完成的相关调用，一个客户端可以被多个 goroutine 使用。
type Client struct {
	seq      uint64           // 编号
	cc       codec.Codec      // 消息的编解码器
	opt      *Option          // 协商编解码方式
	pending  map[uint64]*Call // 存储未完成的消息
	header   codec.Header     // 复用请求头，每个客户端只需要一个请求头
	sending  sync.Mutex       // 互斥锁，为了保证请求的有序发送
	mu       sync.Mutex       // 锁
	closing  bool             // 用户停止调用（用户主动）
	shutdown bool             // 服务停止（出bug）
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

// Close 关闭连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return ErrShutdown
	}
	c.closing = true
	return c.cc.Close()
}

// IsAvailable 客户端是否在工作
func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.shutdown && !c.closing
}

// registerCall 注册客户端
func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

// removeCall 从客户端移除对应的 call 并且返回实例
func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

// terminateCalls 当调用发生错误，通知所有 pending 状态的 Call
func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}

// receive 接收响应
func (c *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		// 读取 header，存储到 Client
		if err = c.cc.ReadHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			// 写部分失效，调用失败
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body" + err.Error())
			}
			call.done()
		}
	}
}

// NewClient 新建Client实例
func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	// 协议交换，GobCodec 赋值给 f
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("无效的编码器类型：%s", opt.CodecType)
		log.Println("rpc client：codec error：", err)
		return nil, err
	}
	// 与服务器一起发送选项，编码存储 opt
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil
}

// newClientCodec 给客户端的设置和编解码器赋值，创建子协程，接收响应
func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, // Seq从1开始，0表示无效调用
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

// parseOptions 解析编解码方式
func parseOptions(opts ...*Option) (*Option, error) {
	// 如果opts为nil或传递nil作为参数，就返回默认的，否则就往下执行
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

// Dial 连接到指定网络地址的RPC服务器
func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	// 如果 client 为空就关闭连接
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}

// send 发送 call
func (client *Client) send(call *Call) {
	// 确保客户端发送完整请求
	client.sending.Lock()
	defer client.sending.Unlock()

	// 注册 call
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// 准备请求头
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	// 发送加密之后的请求
	if err := client.cc.Writer(&client.header, call.Args); err != nil {
		// 发送完请求之后删除等待消息的 call 并获取实例
		call := client.removeCall(seq)
		// call 如果是 nil 就说明写入失败
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

// Go 异步调用函数，返回 Call 结构
func (c *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)
	return call
}

// Call 是对 Go 的封装，阻塞 call.Done，等待响应返回，是一个同步接口。
func (client *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}
