package main

import (
	"github.com/gogf/gf/g"
	"github.com/gogf/gf/g/net/ghttp"
	"github.com/zhaopengme/gf-jwt/example/auth"
)

func hello(r *ghttp.Request) {
	r.Response.Write("哈喽世界！")
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
		r.Response.Write("it's work！")
	})
	s.SetPort(8000)
	s.Run()
}
