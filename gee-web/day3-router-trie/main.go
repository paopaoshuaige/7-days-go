package main

import (
	"net/http"
	"pee"
)

func main() {
	p := pee.New()

	// 调用GET方法添加路由
	p.GET("/", func(c *pee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello,Pee</h1>")
	})

	p.GET("/hello", func(c *pee.Context) {
		// /hello?name=lzj
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	p.GET("/hello/:name", func(c *pee.Context) {
		// expect /hello/geektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})

	p.GET("/assets/*filepath", func(c *pee.Context) {
		c.JSON(http.StatusOK, pee.H{
			"filepath": c.Param("filepath"),
		})
	})

	p.Run(":9999")
}
