package main

import (
	"github.com/gogf/gf-jwt/example/auth"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"time"
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
func MiddlewareAuth(r *ghttp.Request) {
	auth.GfJWTMiddleware.MiddlewareFunc()(r)
	r.Middleware.Next()
}

func MiddlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

func main() {
	println(time.Now().Unix())
	s := g.Server()
	s.BindHandler("/", works)
	s.BindHandler("POST:/login", auth.GfJWTMiddleware.LoginHandler)
	s.Group("/user", func(g *ghttp.RouterGroup) {
		g.Middleware(MiddlewareCORS, MiddlewareAuth)
		g.ALL("/refresh_token", auth.GfJWTMiddleware.RefreshHandler)
		g.ALL("/hello", hello)
	})
	s.SetPort(8000)
	s.Run()
}
