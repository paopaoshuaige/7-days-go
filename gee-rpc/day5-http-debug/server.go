package geerpc

import (
	"day4-timeout/codec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"
)

const MagicNumber = 0x3bef5c

// Option 协商编解码方式，处理超时
type Option struct {
	MagicNumber    int           // 用来标记是 geerpc 请求
	CodecType      string        // 客户可以选择不同的编解码器来编码主体
	ConnectTimeout time.Duration // 连接超时
	HandleTimeout  time.Duration // 处理超时
}

// DefaultOption 默认的编解码方式，为了简单使用固定的 JSON 编码格式
var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10,
}

// request 保存通话的所有信息
type request struct {
	h            *codec.Header // 请求头
	argv, replyv reflect.Value // 请求参数和响应数据
	mtype        *methodType
	svc          *service
}

// Server RPC 服务端
type Server struct {
	serviceMap sync.Map
}

// NewServer 返回一个新的服务端
func NewServer() *Server {
	return &Server{}
}

// DefaultServer Server 的默认实例
var DefaultServer = NewServer()

// Accept 接收监听器上的每个传入的连接和服务请求
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

// Register 注册默认服务
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

// Register 服务端方法注册
func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

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
	server.serveCodec(f(conn), &opt)
}

// invalidRequest 是一个占位符
var invalidRequest = struct{}{}

// serveCodec 通过编解码器处理消息
func (server *Server) serveCodec(cc codec.Codec, opt *Option) {
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
		go server.handleRequest(cc, req, sending, wg, opt.HandleTimeout)
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
	req.svc, req.mtype, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	// 创建两个入参实例
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	// 确保argvi是一个指针，ReadBody需要一个指针作为参数
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	// 把请求报文反序列化为第一个入参 argv
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
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
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	// 调用注册的 rpc 方法来获得正确的参数，打印参数并发送 hello 消息，锁-1
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	// 这里需要确保 sendResponse 仅调用一次
	// called 信道接收到消息，代表处理没有超时，继续执行 sendResponse。
	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			// 完成方法调用
			server.sendResponse(cc, req.h, invalidRequest, sending)
			return
		}
		server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()
	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	// time.After() 先于 called 接收到消息，说明处理已经超时，called 和 sent 都将被阻塞。在此处调用 sendResponse
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
}

// findService 查找服务
func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	// 分割成服务名和方法名
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	// 找到map存储的服务名对应的实例
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	// 断言
	svc = svci.(*service)
	// 通过方法名找到方法类型
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}
