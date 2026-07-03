// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"
	"strings"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveIAMUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 移除用户的某个角色；使用 query：?user_code=&role_code=
func NewRemoveIAMUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveIAMUserRoleLogic {
	return &RemoveIAMUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveIAMUserRoleLogic) RemoveIAMUserRole(req *types.RemoveIAMUserRoleReq) (resp *types.RemoveIAMUserRoleReply, err error) {
	userCode := strings.TrimSpace(req.UserCode)
	roleCode := strings.TrimSpace(req.RoleCode)
	if userCode == "" || roleCode == "" {
		return nil, errInvalid("user_code 与 role_code 均不能为空")
	}
	if err := l.svcCtx.Rbac.RemoveUserRole(userCode, roleCode); err != nil {
		return nil, errInvalid(err.Error())
	}
	if err := svc.EnsureUserCasbinGroupings(l.svcCtx.Rbac, l.svcCtx.Casbin, userCode); err != nil {
		return nil, errInvalid("同步 Casbin 用户角色失败")
	}
	return &types.RemoveIAMUserRoleReply{}, nil
}
