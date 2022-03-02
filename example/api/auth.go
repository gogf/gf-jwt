package api

import (
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

type AuthLoginReq struct {
	g.Meta `path:"/login" method:"post"`
}

type AuthLoginRes struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

type AuthRefreshTokenReq struct {
	g.Meta `path:"/refresh_token" method:"post"`
}

type AuthRefreshTokenRes struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

type AuthLogoutReq struct {
	g.Meta `path:"/logout" method:"post"`
}

type AuthLogoutRes struct {
}
