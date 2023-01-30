package pee

import (
	"fmt"
	"log"
	"net/http"
)

// 将HandlerFunc定义为此func
type HandlerFunc func(http.ResponseWriter, *http.Request)

type Engine struct {
	// 路由映射表
	// key由静态方法和静态路由地址构成，如GET-/、GET-/hello、POST-/hello
	// 相同的路由不同的请求方法可以映射到不同的处理方法(Handler)
	// value是用户映射的处理方法
	router map[string]HandlerFunc
}

func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

// 把路由和请求方法注册到映射表router
func (e *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	log.Printf("Router %4s - %s", method, pattern)
	e.router[key] = handler
}

// GET请求
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRoute("GET", pattern, handler)
}

// POST
func (e *Engine) POST(patter string, handler HandlerFunc) {
	e.addRoute("POST", patter, handler)
}

// ListenAndServe的包装，启动httpserver
func (e *Engine) Run(add string) (err error) {
	// ListenAndServe方法里面会去调用 handler.ServeHTTP()方法
	return http.ListenAndServe(add, e)
}

// 实现Handler接口中的ServeHTTP方法
// 解析请求的路径，查找路由映射表，如果查到，就执行注册的处理方法。
// 如果查不到，就返回 404 NOT FOUND。
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.Method + "-" + req.URL.Path
	if handler, ok := e.router[key]; ok {
		// 这里调用的实际上是main函数中传进来的func方法，我们重写一下，如下
		handler(w, req)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 not found:%s\n", req.URL)
	}
}
