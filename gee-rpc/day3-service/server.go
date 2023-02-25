package geerpc

import (
	"day2-client/codec"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x3bef5c

// Option 协商编解码方式
type Option struct {
	MagicNumber int    // 用来标记是 geerpc 请求
	CodecType   string // 客户可以选择不同的编解码器来编码主体
}

// DefaultOption 默认的编解码方式，为了简单使用固定的 JSON 编码格式
var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

// request 保存通话的所有信息
type request struct {
	h            *codec.Header // 请求头
	argv, replyv reflect.Value // 请求参数和响应数据
}

// Server RPC 服务端
type Server struct{}

// NewServer 返回一个新的服务端
func NewServer() *Server {
	return &Server{}
}

// DefaultServer Server 的默认实例
var DefaultServer = NewServer()

// Accept 接收监听器上的每个传入的连接和服务请求
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

// Accept 接收监听器上的每个传入的连接和服务请求
func (server *Server) Accept(lis net.Listener) {
	// 等待 socket 连接建立并开启子协程处理
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go server.ServeConn(conn)
	}
}

// ServeConn 在单个链接上运行并且一直阻塞提供服务，直到客户端挂断
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option
	// 反序列化得到 Option 实例
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	// 检查 MagicNumber 和 CodeType 的值是否正确
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	// 根据 CodeType 得到对应的消息编解码器
	server.serveCodec(f(conn))
}

// invalidRequest 是一个占位符
var invalidRequest = struct{}{}

// serveCodec 通过编解码器处理消息
func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // 确保发送完整的响应
	wg := new(sync.WaitGroup)  // 等待所有请求处理完毕
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break // 这是不可能恢复的，所以关闭连接
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		// 加一个锁
		wg.Add(1)
		// 使用了协程并发执行请求。
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

// readRequestHeader 读取请求头
func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

// readRequest 读取请求
func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	// 我们不知道参数类型，假设他是string，给参数 body 解码
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	return req, nil
}

// sendResponse 存储序列化之后的 Header 和 Body，回复请求
func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Writer(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

// handleRequest 处理请求，返回结果
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	// 调用注册的 rpc 方法来获得正确的参数，打印参数并发送 hello 消息，锁-1
	defer wg.Done()
	// 打印请求头和参数
	log.Println(req.h, req.argv.Elem())
	// 请求头的序列号写入 replyv
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
