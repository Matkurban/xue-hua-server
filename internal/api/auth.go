package api

import (
	"server/v3/pkg/common/a2r"
	"github.com/gin-gonic/gin"
	"github.com/openimsdk/protocol/auth"
)

type AuthApi struct {
	Client auth.AuthClient
}

func NewAuthApi(client auth.AuthClient) AuthApi {
	return AuthApi{client}
}

func (o *AuthApi) GetAdminToken(c *gin.Context) {
	a2r.Call(c, auth.AuthClient.GetAdminToken, o.Client)
}

func (o *AuthApi) GetUserToken(c *gin.Context) {
	a2r.Call(c, auth.AuthClient.GetUserToken, o.Client)
}

func (o *AuthApi) ParseToken(c *gin.Context) {
	a2r.Call(c, auth.AuthClient.ParseToken, o.Client)
}

func (o *AuthApi) ForceLogout(c *gin.Context) {
	a2r.Call(c, auth.AuthClient.ForceLogout, o.Client)
}
