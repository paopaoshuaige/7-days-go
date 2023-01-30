package main

import (
	"log"
	"net/http"
	"pee"
	"time"
)

func onlyForV2() pee.HandlerFunc {
	return func(c *pee.Context) {
		// Start timer
		t := time.Now()
		// 记录一下错误服务（这里默认设置都是500）
		c.Fail(500, "Internal Server Error")
		// 计算用时
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func main() {
	r := pee.New()
	r.Use(pee.Logger()) // global midlleware
	r.GET("/", func(c *pee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})

	v2 := r.Group("/v2")
	// 先执行handler，然后logger中间件的next之前，然后执行only，执行完之后执行logger
	v2.Use(onlyForV2()) // v2 group middleware
	{
		v2.GET("/hello/:name", func(c *pee.Context) {
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}

	r.Run(":9999")
}
