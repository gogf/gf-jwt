package main

import (
	"github.com/gogf/gf-jwt/example/auth"
	"github.com/gogf/gf/g"
	"github.com/gogf/gf/g/net/ghttp"
)

func hello(r *ghttp.Request) {
	r.Response.Write("Hello World!")
}

func main() {
	s := g.Server()
	a := new(auth.Default)
	a.Init()

	s.BindHandler("POST:/login", a.GfJWTMiddleware.LoginHandler)

	s.Group("/user").Bind("/user", []ghttp.GroupItem{
		{"ALL", "*", func(r *ghttp.Request) {
			r.Response.CORSDefault()
			a.GfJWTMiddleware.MiddlewareFunc()(r)
		}, ghttp.HOOK_BEFORE_SERVE},
		{"GET", "/refresh_token", a.GfJWTMiddleware.RefreshHandler},
		{"GET", "/hello", hello},
	})

	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.Write("It works！")
	})
	s.SetPort(8000)
	s.Run()
}
