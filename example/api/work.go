package api

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var Work = new(workApi)

type workApi struct{}

// Works works is the default router handler for web server.
func (a *workApi) Works(r *ghttp.Request) {
	data := g.Map{
		"message": "It works!",
	}
	r.Response.WriteJson(data)
}

// info should be authenticated to view.
// info is the get user data handler
func (a *workApi) Info(r *ghttp.Request) {
	data := g.Map{
		// get identity by identity key 'id'
		"id":           r.Get("id"),
		"identity_key": r.Get(Auth.IdentityKey),
		// get payload by identity
		"payload": r.Get("JWT_PAYLOAD"),
	}
	r.Response.WriteJson(data)
}
