package main

import (
	"net/http"
	"pee"
)

func main() {
	r := pee.Default()
	r.GET("/", func(c *pee.Context) {
		c.String(http.StatusOK, "Hello lzj\n")
	})
	// index out of range for testing Recovery()
	r.GET("/panic", func(c *pee.Context) {
		names := []string{"lzj"}
		c.String(http.StatusOK, names[100])
	})

	r.Run(":9999")
}
