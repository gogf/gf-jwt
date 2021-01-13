package service

import (
	"github.com/gogf/gf-jwt/example/model"
	"github.com/gogf/gf/frame/g"
)

var User = new(userService)

type userService struct {}

func (s *userService) GetUserByUsernamePassword(serviceReq *model.ServiceLoginReq) map[string]interface{} {
	if serviceReq.Username == "admin" && serviceReq.Password == "admin" {
		return g.Map{
			"id": 1,
			"username": "admin",
		}
	}
	return nil
}
