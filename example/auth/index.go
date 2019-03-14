package auth

import (
	"github.com/gogf/gf/g"
	"github.com/gogf/gf/g/net/ghttp"
	"github.com/gogf/gf/g/util/gvalid"
	"github.com/zhaopengme/gf-jwt"
	"log"
	"net/http"
	"time"
)

type Default struct {
	GfJWTMiddleware *jwt.GfJWTMiddleware
	Rules            map[string]string
}

func (d *Default) Init() {
	authMiddleware, err := jwt.New(&jwt.GfJWTMiddleware{
		Realm:           "test zone",
		Key:             []byte("secret key"),
		Timeout:         time.Minute * 5,
		MaxRefresh:      time.Minute * 5,
		IdentityKey:     "id",
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
		Authenticator:   d.Authenticator,
		LoginResponse:   d.LoginResponse,
		RefreshResponse: d.RefreshResponse,
		Unauthorized:    d.Unauthorized,
		IdentityHandler: d.IdentityHandler,
		PayloadFunc:     d.PayloadFunc,
	})
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	d.GfJWTMiddleware = authMiddleware
	d.Rules = map[string]string{
		"username": "required",
		"password": "required",
	}
}

func (d *Default) PayloadFunc(data interface{}) jwt.MapClaims {
	claims := jwt.MapClaims{}
	params := data.(map[string]interface{})
	if len(params) > 0 {
		for k, v := range params {
			claims[k] = v
		}
	}
	return claims
}

func (d *Default) IdentityHandler(r *ghttp.Request) interface{} {
	claims := jwt.ExtractClaims(r)
	return claims["id"]
}

func (d *Default) Unauthorized(r *ghttp.Request, code int, message string) {
	r.Response.WriteJson(g.Map{
		"code": code,
		"msg":  message,
	})
	r.ExitAll()
}

func (d *Default) LoginResponse(r *ghttp.Request, code int, token string, expire time.Time) {
	r.Response.WriteJson(g.Map{
		"code":   http.StatusOK,
		"token":  token,
		"expire": expire.Format(time.RFC3339),
	})
	r.ExitAll()
}

func (d *Default) RefreshResponse(r *ghttp.Request, code int, token string, expire time.Time) {
	r.Response.WriteJson(g.Map{
		"code":   http.StatusOK,
		"token":  token,
		"expire": expire.Format(time.RFC3339),
	})
	r.ExitAll()
}

func (d *Default) Authenticator(r *ghttp.Request) (interface{}, error) {
	data := r.GetMap()
	if e := gvalid.CheckMap(data, d.Rules); e != nil {
		return "", jwt.ErrFailedAuthentication
	}
	if (data["username"] == "admin" && data["password"] == "admin") {
		return g.Map{
			"username": data["username"],
			"id":       data["username"],
		}, nil
	}

	return nil, jwt.ErrFailedAuthentication
}
