package pee

import (
	"log"
	"net/http"
)

// 将HandlerFunc定义为此func，使用了context
type HandlerFunc func(*Context)

type (
	Engine struct {
		// 路由映射表，key由静态方法和静态路由地址构成，如GET-/、GET-/hello、POST-/hello
		// 相同的路由不同的请求方法可以映射到不同的处理方法(Handler)，value是用户映射的处理方法
		router       *router // router.go里面的结构体
		*RouterGroup         // 这是 go 中的嵌套类型，类似 Java/Python 等语言的继承。这样 Engine 就可以拥有 RouterGroup 的属性了。
		groups       []*RouterGroup
	}

	RouterGroup struct {
		prifix      string        // 前缀
		middlewares []HandlerFunc // 中间件
		engine      *Engine       // 分组由他控制
	}
)

// 给en的分组和组赋值，Group里面的engine里面的Group和Groups是一个，地址一样。
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// 所有组共享同一个引擎实例
func (g *RouterGroup) Group(prefix string) *RouterGroup {
	engine := g.engine
	newGroup := &RouterGroup{ // 每次进来都新建一个路由组保存分组前缀。
		prifix: g.prifix + prefix,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup) // 把当前组加入到分组控制路由的组里
	return newGroup
}

// 把路由和请求方法注册到映射表router
func (g *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := g.prifix + comp
	log.Printf("Route %4s - %s", method, pattern)
	g.engine.router.addRoute(method, pattern, handler)
}

// GET请求
func (g *RouterGroup) GET(pattern string, handler HandlerFunc) {
	g.addRoute("GET", pattern, handler)
}

// POST
func (g *RouterGroup) POST(patter string, handler HandlerFunc) {
	g.addRoute("POST", patter, handler)
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
	c := newContext(w, req)
	e.router.handle(c)
}
