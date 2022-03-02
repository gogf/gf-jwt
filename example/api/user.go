package api

import "github.com/gogf/gf/v2/frame/g"

type UserGetInfoReq struct {
	g.Meta `path:"/user/info" method:"get"`
}

type UserGetInfoRes struct {
	Id          int    `json:"id"`
	IdentityKey string `json:"identity_key"`
	Payload     string `json:"payload"`
}
