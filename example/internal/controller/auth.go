package controller

import (
	"context"
	"github.com/gogf/gf-jwt/v2/example/api"
	"github.com/gogf/gf-jwt/v2/example/internal/service"
)

type authController struct{}

var Auth = authController{}

func (c *authController) Login(ctx context.Context, req *api.AuthLoginReq) (res *api.AuthLoginRes, err error) {
	res = &api.AuthLoginRes{}
	res.Token, res.Expire = service.Auth().LoginHandler(ctx)
	return
}

func (c *authController) RefreshToken(ctx context.Context, req *api.AuthRefreshTokenReq) (res *api.AuthRefreshTokenRes, err error) {
	res = &api.AuthRefreshTokenRes{}
	res.Token, res.Expire = service.Auth().RefreshHandler(ctx)
	return
}

func (c *authController) Logout(ctx context.Context, req *api.AuthLogoutReq) (res *api.AuthLogoutRes, err error) {
	service.Auth().LogoutHandler(ctx)
	return
}
