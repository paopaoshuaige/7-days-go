package main

import (
	"fmt"
	"html/template"
	"net/http"
	"pee"
	"time"
)

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := pee.New()
	r.Use(pee.Logger())
	r.SetFuncMap(template.FuncMap{ // 这里的自定义的函数可以在html里面调用，key表示在模板中使用的函数名，value是对应的实现函数。
		"FormatAsDate": FormatAsDate,
	})
	// templates文件夹下的全部资源都载入内存
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{Name: "paopao", Age: 20}
	stu2 := &student{Name: "lzj", Age: 22}

	r.GET("/", func(c *pee.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})

	r.GET("/students", func(c *pee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", pee.H{
			"title":  "pee",
			"stuArr": [2]*student{stu1, stu2},
		})
	})

	r.GET("/date", func(c *pee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", pee.H{
			"title": "pee",
			"now":   time.Now(),
		})
	})

	r.Run(":9999")
}
