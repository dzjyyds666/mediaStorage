package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dzjyyds666/Allspark-go/jwtx"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/vortex/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type LoginHandler struct {
	ctx context.Context
}

// 签名token
func (lh *LoginHandler) HandleLogin(ctx *vortex.Context) error {
	var req loginReq
	err := json.NewDecoder(ctx.Request().Body).Decode(&req)
	if nil != err {
		logx.Errorf("StorageServer|HandleLogin|decode login req error: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.ParamsInvaild, echo.Map{
			"msg": "参数错误",
		})
	}

	if req.UserName != s.admin.Username || req.Password != s.admin.Password {
		logx.Errorf("StorageServer|HandleLogin|username: %s|password: %s", req.UserName, req.Password)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.ParamsInvaild, echo.Map{
			"msg": "用户名或密码错误",
		})
	}

	jwtToken, err := jwtx.SignJwt(s.jwtToken.Secret, jwt.MapClaims{
		"uid":     req.UserName,
		"expires": time.Now().Add(time.Duration(s.jwtToken.Expire) * time.Second).Unix(),
	})
	if err != nil {
		logx.Errorf("StorageServer|HandleLogin|SignJwt|err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError, echo.Map{
			"msg": "登录失败",
		})
	}

	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"msg":  "登录成功",
		"jwt":  jwtToken,
		"user": req.UserName,
	})
}
