package main

import (
	"encoding/json"
	"fmt"
	"geerpc"
	"geerpc/codec"
	"log"
	"net"
	"time"
)

// startServer 启动服务
func startServer(addr chan string) {
	// 选择一个没有使用的端口
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalln("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	// 使用了信道 addr 确保服务端端口监听成功再发起请求
	addr <- l.Addr().String()
	geerpc.Accept(l)
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)

	// rpc client
	// 客户端首先发送 Option 进行协议交换，接下来发送消息头，消息体
	conn, _ := net.Dial("tcp", <-addr)
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)
	// 发送，创建一个新的解码链接
	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
	// 初始化该链接
	cc := codec.NewGobCodec(conn)
	// 发送请求，接收响应
	for i := 0; i < 5; i++ {
		// 请求头
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		// 写入请求头
		_ = cc.Writer(h, fmt.Sprintf("geerpc req %d", h.Seq))
		// 读取请求头
		_ = cc.ReadHeader(h)
		var reply string
		// 读取 body
		_ = cc.ReadBody(&reply)
		// 打印body
		log.Println("reply:", reply)
	}
}
