# gf-jwt
GF jwt plugin

This plugin is forked [https://github.com/appleboy/gin-jwt](https://github.com/appleboy/gin-jwt) plugin, modified to [https://github.com/gogf/gf](https://github.com/gogf/gf) plugin.


[英文](README.md) [中文](README_zh.md)


## Use

Download and install

```sh
$ go get github.com/gogf/gf-jwt
```

Import

```go
import "github.com/gogf/gf-jwt"
```

## Demo

Run `example/main.go` on the `8000` port.

```bash
$ go run example/main.go
```

![api screenshot](screenshot/server.png)

Test the effect on the command line via [httpie](https://github.com/jkbrzt/httpie) or curl.

### Login interface:

```bash
$ http -v --form POST localhost:8000/login username=admin password=admin
```
or
```bash
$ curl -X POST -d 'username=admin&password=admin' localhost:8000/login
```

Command line output

![api screenshot](screenshot/login.png)

### Refresh token interface:

```bash
$ http -v -f GET localhost:8000/refresh_token "Authorization:Bearer xxxxxxxxx" "Content-Type: application/json"
```
or
```bash
$ curl -H 'Authorization:Bearer xxxxxxxxx' -X POST localhost:8000/refresh_token
```

Command line output

![api screenshot](screenshot/refresh_token.png)

### User info interface

We test the return of the info interface with the username `admin` and password `admin`

```bash
$ http -f GET localhost:8000/user/info "Authorization:Bearer xxxxxxxxx" "Content-Type: application/json"
```
or
```bash
$ curl -H 'Authorization:Bearer xxxxxx' -X POST localhost:8000/user/info
```

Command line output

![api screenshot](screenshot/hello.png)


Thanks again [https://github.com/appleboy/gin-jwt](https://github.com/appleboy/gin-jwt)
