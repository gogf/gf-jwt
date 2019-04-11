package main

import (
	"github.com/gogf/gf-jwt/example/auth"
	"github.com/gogf/gf/g"
	"github.com/gogf/gf/g/net/ghttp"
)

// hello should be authenticated to view.
func hello(r *ghttp.Request) {
	r.Response.Write("Hello World!")
}

// works is the default router handler for web server.
func works(r *ghttp.Request) {
	r.Response.Write("It works!")
}

// authHook is the HOOK function implements JWT logistics.
func authHook(r *ghttp.Request) {
	r.Response.CORSDefault()
	auth.GfJWTMiddleware.MiddlewareFunc()(r)
}

func main() {
	s := g.Server()
	s.Group().Bind("/", []ghttp.GroupItem{
		{"ALL",  "/",             works},
		{"POST", "/login",        auth.GfJWTMiddleware.LoginHandler},
	})
	s.Group("/user").Bind("/user", []ghttp.GroupItem{
		{"ALL", "*",              authHook, ghttp.HOOK_BEFORE_SERVE},
		{"GET", "/refresh_token", auth.GfJWTMiddleware.RefreshHandler},
		{"GET", "/hello",         hello},
	})
	s.SetPort(8000)
	s.Run()
}
