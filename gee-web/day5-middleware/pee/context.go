package pee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

// 上下文结构体
type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request
	// request信息
	Path   string
	Method string
	Params map[string]string
	// response信息
	StatusCode int
	// 中间件
	handlers []HandlerFunc
	index    int
}

// 获取Params对应的value
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// 返回一个ServerHTTP的上下文
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Path:   req.URL.Path,
		Method: req.Method,
		Req:    req,
		Writer: w,
		index:  -1,
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	// 这里是
	c.JSON(code, H{"message": err})
}

func (c *Context) Next() {
	c.index++ // 当前执行到第几个中间件
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// 查询url路径中key对应的value，Post形式
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// 获取url路径中key对应的字符串，Get形式
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// 写入http状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// 写入请求头
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// 返回一个字符串结果
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// 写入json格式数据
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// 写入数据
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// 把字符出转换成HTML
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
