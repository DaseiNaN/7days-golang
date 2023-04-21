package main

import (
	"example/gee"
	"log"
	"net/http"
)

func main() {
	r := gee.New()
	r.Get("/index", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	v1 := r.Group("v1")
	{
		v1.Get("/", func(c *gee.Context) {
			c.String(http.StatusOK, "<h1>V1 Root</h1>")
		})

		v1.Get("/hello", func(c *gee.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}

	v2 := r.Group("v2")
	{
		v2.Get("/", func(c *gee.Context) {
			c.String(http.StatusOK, "<h1>V2 Root</h1>")
		})

		v2.Get("/hello/:name", func(c *gee.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *gee.Context) {
			// curl "http://localhost:9999/v2/login" -X POST -d 'username=dasein&password=OC1234'
			c.JSON(http.StatusOK, gee.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
	}
	log.Fatal(r.Run(":9999"))
}
