// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// 开发环境默认账号（仅用于 Swagger/联调，生产应改为真实登录或配置）
const devUsername = "admin"
const devPassword = "admin"

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 登录获取 JWT，用于 Swagger Authorize 或前端
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginReply, error) {
	if req.Username != devUsername || req.Password != devPassword {
		return nil, errors.New("用户名或密码错误")
	}

	secret := l.svcCtx.Config.Auth.AccessSecret
	expire := l.svcCtx.Config.Auth.AccessExpire
	if expire <= 0 {
		expire = 86400
	}
	iat := time.Now().Unix()
	exp := iat + expire

	claims := jwt.MapClaims{
		"exp": exp,
		"iat": iat,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &types.LoginReply{
		AccessToken: tokenStr,
		Expire:      exp,
	}, nil
}
