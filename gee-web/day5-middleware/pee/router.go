package pee

import (
	"net/http"
	"strings"
)

// 路由表结构体，roots存储请求方式Trie的根节点
type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

// roots key eg, roots['GET'] roots['POST']
// handlers key eg, handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']

// 新建一个路由映射表
func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// 只允许一个*
func parsePatten(pattern string) []string {
	// 按/分割字符串，/会被去掉，然后会多一个""在0下标
	vs := strings.Split(pattern, "/")

	// 存储url后缀，vs会多一个""
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

// 把映射加入路由表，method是请求方法，pattern是路径，handler是路由表信息
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePatten(pattern)

	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		// 如果没有这个方法对应的根，那就创建一个
		r.roots[method] = &node{}
	}
	// 然后插入该路径
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

// 获取全部的路由
func (r *router) getRouter(method string, path string) (*node, map[string]string) {
	searchParts := parsePatten(path) // 分割一下url
	params := make(map[string]string)
	root, ok := r.roots[method] // 获取当前请求方法的树根

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0) // 搜索对应的节点

	if n != nil { // 如果找到对应的节点了
		parts := parsePatten(n.pattern) // 获取剩下的待匹配路由
		for i, part := range parts {
			if part[0] == ':' { // 如果该段该匹配的路由有：
				// 因为有冒号，所以什么都能匹配，直接把对应的映射上就好了
				// 比如/hello/:name /hello/lzj
				// 映射出来就是params[name] = lzj
				params[part[1:]] = searchParts[i]
			}
			if part[0] == '*' && len(part) > 1 { // 同上
				params[part[1:]] = strings.Join(searchParts[i:], "/")
				break
			}
		}
		return n, params // 返回当前节点和映射的map
	}
	return nil, nil
}

// 解析路由映射表，然后给对应的handler方法传入当前ServeHTTP上下文
func (r *router) handle(c *Context) {
	// 先获取节点和路由
	n, params := r.getRouter(c.Method, c.Path)
	if n != nil {
		// 把获取到的路由映射绑定到上下文
		c.Params = params
		// 此时的key就是上下文中的请求方法和n的待匹配url
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key]) // 传过来的上下文里面有刚才配到的中间件
		// 然后加入路由请求的handler函数
	} else {
		// 否则把报错的handler加入进去，运行到这里的时候就会报错了
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
