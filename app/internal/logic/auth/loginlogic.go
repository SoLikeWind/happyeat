// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// devUsername 开发态内置超管：用户名 admin 映射到 user_code=dev-admin（密码存库，默认 admin123）。
const devUsername = "admin"

var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

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

func loginUserCode(username string) string {
	u := strings.TrimSpace(username)
	if u == "" {
		return ""
	}
	if strings.EqualFold(u, devUsername) {
		return "dev-admin"
	}
	return strings.ToLower(u)
}

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginReply, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}

	// 账号即手机号：11 位手机号走手机号登录，其它（如 admin）走 user_code 兼容登录。
	var (
		user *svc.LoginUser
		err  error
	)
	if phoneRe.MatchString(username) {
		user, err = l.svcCtx.Rbac.AuthenticateByPhone(username, req.Password)
	} else {
		user, err = l.svcCtx.Rbac.AuthenticateByUserCode(loginUserCode(username), req.Password)
	}
	if err != nil {
		return nil, err
	}

	primaryRole := svc.PickPrimaryRole(user.Roles)
	secret := l.svcCtx.Config.Auth.AccessSecret
	expire := l.svcCtx.Config.Auth.AccessExpire
	if expire <= 0 {
		expire = 86400
	}
	iat := time.Now().Unix()
	exp := iat + expire

	// user_code：Casbin 主体；role：非标准 claim，供前端解析。
	claims := jwt.MapClaims{
		"exp":       exp,
		"iat":       iat,
		"sub":       user.UserCode,
		"user_code": user.UserCode,
		"role":      primaryRole,
		"roles":     user.Roles,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &types.LoginReply{
		AccessToken: tokenStr,
		Expire:      exp,
		UserCode:    user.UserCode,
		Role:        primaryRole,
		Roles:       user.Roles,
		RoleNames:   user.RoleNames,
	}, nil
}
