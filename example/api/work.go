package api

import (
	"github.com/gogf/gf/net/ghttp"
)

var Work = new(workApi)

type workApi struct{}

// hello should be authenticated to view.
func (a *workApi) Hello(r *ghttp.Request) {
	r.Response.Write("Hello")
}

// works is the default router handler for web server.
func (a *workApi) Works(r *ghttp.Request) {
	r.Response.Write("It works!")
}
