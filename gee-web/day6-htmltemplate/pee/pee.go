package pee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

// 将HandlerFunc定义为此func，使用了context
type HandlerFunc func(*Context)

type (
	Engine struct {
		// 路由映射表，key由静态方法和静态路由地址构成，如GET-/、GET-/hello、POST-/hello
		// 相同的路由不同的请求方法可以映射到不同的处理方法(Handler)，value是用户映射的处理方法
		router        *router // router.go里面的结构体
		*RouterGroup          // 这是 go 中的嵌套类型，类似 Java/Python 等语言的继承。这样 Engine 就可以拥有 RouterGroup 的属性了。
		groups        []*RouterGroup
		htmlTemplates *template.Template
		funcMap       template.FuncMap
	}

	RouterGroup struct {
		prefix      string        // 前缀
		middlewares []HandlerFunc // 中间件
		engine      *Engine       // 分组由他控制
	}
)

// 存储自定义模板映射
func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

// 模板加载进内存
func (e *Engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}

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
		prefix: g.prefix + prefix,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup) // 把当前组加入到分组控制路由的组里
	return newGroup
}

// 把路由和请求方法注册到映射表router
func (g *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := g.prefix + comp
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

// 加入中间件
func (g *RouterGroup) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// 创建静态handler
func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	// 获取当前路由组前缀拼接
	absolutePath := path.Join(g.prefix, relativePath)
	// httpfileserver返回一个Handler，这个Handler向httpreq提供位于上面代码root变量的文件系统的内容
	// 直接定位到这个目录下的index.html文件。root一般使用http.Dir(“yourFilePath”)
	// StripPrefix将URL中的前缀中的第一个参数字符串删除，然后再交给后面的Handler处理
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath") // 查找url上filepath对应的参数
		// 检查文件是否存在，是否有权限访问
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// 注册静态handler
func (g *RouterGroup) Static(relativePath string, root string) {
	// 第一个参数是用户指定的路径，第二个是要匹配的文件路径，http.Dir是把字符串转换成html实体码
	handler := g.createStaticHandler(relativePath, http.Dir(root))
	// 拼接路径，/.../*filepath，这里用*就代表可以匹配任意的后缀
	// *filepath 代表贪心匹配，例如 /css/xxx.css，可以匹配剩余的所有子路径。/:filepath 只匹配一层路径。
	urlPattern := path.Join(relativePath, "/*filepath")
	// 注册GEThandler
	g.GET(urlPattern, handler)
}

// 实现Handler接口中的ServeHTTP方法
// 解析请求的路径，查找路由映射表，如果查到，就执行注册的处理方法。
// 如果查不到，就返回 404 NOT FOUND。
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) { // 判断前缀，看请求适合哪个中间件
			middlewares = append(middlewares, group.middlewares...) // 保存适用的中间件
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares // 把需要运行的中间件保存在handlers上去执行。
	c.engine = e             // 为了让模板能用上en指针赋值
	e.router.handle(c)
}
