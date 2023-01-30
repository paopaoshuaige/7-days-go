package main

import (
	"fmt"
	"net/http"
	"pee"
)

func main() {
	p := pee.New()

	// 调用GET方法添加路由
	p.GET("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})

	p.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})

	p.Run(":9999")
}
