package pee

import "net/http"

// 路由表结构体
type router struct {
	handlers map[string]HandlerFunc
}

// 新建一个路由映射表
func newRouter() *router {
	return &router{
		handlers: make(map[string]HandlerFunc),
	}
}

// 把映射加入路由表，method是请求方法，pattern是路径，handler是路由表信息
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	r.handlers[key] = handler
}

// 解析路由映射表，然后给对应的handler方法传入当前ServeHTTP上下文
func (r *router) handle(c *Context) {
	key := c.Method + "-" + c.Path
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
