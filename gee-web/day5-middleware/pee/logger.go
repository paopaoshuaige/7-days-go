package pee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		// 开始时间
		t := time.Now()
		// 处理请求
		c.Next()
		// 计算解决时间
		log.Printf("time: [%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
