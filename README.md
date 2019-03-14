# gf-jwt
Gf jwt plugin

This plugin is forked [https://github.com/appleboy/gin-jwt](https://github.com/appleboy/gin-jwt) plugin, modified to [https://github.com/gogf/ Gf](https://github.com/gogf/gf) plugin.


[英文](README.md) [中文](README_zh.md)


## Use

Download and install

```sh
$ go get github.com/zhaopengme/gf-jwt
```

Import

```go
Import "github.com/zhaopengme/gf-jwt"
```

## example

Check [demo](example/auth/index.go) and use `ExtractClaims` to customize user data.

[embedmd]:# (example/auth/index.go go)

```go
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
	GinJWTMiddleware *jwt.GinJWTMiddleware
	Rules            map[string]string
}

func (d *Default) Init() {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
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
	d.GinJWTMiddleware = authMiddleware
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

```

## Demo

Run `example/server/index.go` on the `8000` port.

```bash
$ go run example/server/index.go
```

![api screenshot](screenshot/server.png)

Test the effect on the command line via [httpie](https://github.com/jkbrzt/httpie).

### Login interface:

```bash
$ http -v --form POST localhost:8000/login username=admin password=admin
```

Command line output

![api screenshot](screenshot/login.png)

### Refresh token interface:

```bash
$ http -v -f GET localhost:8000/user/refresh_token "Authorization:Bearer xxxxxxxxx" "Content-Type: application/json"
```

Command line output

![api screenshot](screenshot/refresh_token.png)

### hello interface

We test the return of the hello interface with the username `admin` and password `admin`

```bash
$ http -f GET localhost:8000/user/hello "Authorization:Bearer xxxxxxxxx" "Content-Type: application/json"
```

Command line output

![api screenshot](screenshot/hello.png)

### User Authentication Interface

We use an unauthorized token to test the return of the hello interface.

```bash
$ http -f GET localhost:8000/user/hello "Authorization:Bearer xxxxxxxxx" "Content-Type: application/json"
```

Command line output

![api screenshot](screenshot/401.png)


Thanks again [https://github.com/appleboy/gin-jwt](https://github.com/appleboy/gin-jwt)