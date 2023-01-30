package main

import (
	"fmt"
	"log"
	"net/http"
)

// 接收所有的HTTP请求
type Engine struct{}

// 参数一：ResponseWriter，包含Header，Writer，WriterHeader方法
// 参数二：Request，该对象包含了该HTTP请求的所有的信息，比如请求地址、Header和Body等信息
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	case "/hello":
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	default:
		fmt.Fprintf(w, "404 not found:%s\n", req.URL)
	}
}

func main() {
	// 这里用new创建的是一个指针类型结构体
	engine := new(Engine)
	// 所有操作都给engine
	log.Fatal(http.ListenAndServe(":8080", engine))
}
