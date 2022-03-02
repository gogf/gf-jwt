package cmd

import (
	"context"
	"github.com/gogf/gf-jwt/v2/example/internal/controller"
	"github.com/gogf/gf-jwt/v2/example/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server of simple gf-jwt demos",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				// Group middlewares.
				group.Middleware(
					service.Middleware().CORS,
					ghttp.MiddlewareHandlerResponse,
				)
				// Register route handlers.
				group.Bind(
					controller.Auth,
				)
				// Special handler that needs authentication.
				group.Group("/", func(group *ghttp.RouterGroup) {
					group.Middleware(service.Middleware().Auth)
					group.ALLMap(g.Map{
						"/user/info": controller.User.Info,
					})
				})
			})
			// Just run the server.
			s.SetPort(8199)
			s.Run()
			return nil
		},
	}
)
