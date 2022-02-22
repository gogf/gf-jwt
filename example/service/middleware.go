package service

import "github.com/gogf/gf/v2/net/ghttp"

var Middleware = new(middlewareService)

type middlewareService struct{}

func (s *middlewareService) CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}
