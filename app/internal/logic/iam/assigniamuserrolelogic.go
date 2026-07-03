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

type AssignIAMUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 为用户绑定角色（幂等：已绑定则成功）；path 固定便于 Casbin obj 精确匹配
func NewAssignIAMUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignIAMUserRoleLogic {
	return &AssignIAMUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignIAMUserRoleLogic) AssignIAMUserRole(req *types.AssignIAMUserRoleReq) (resp *types.AssignIAMUserRoleReply, err error) {
	userCode := strings.TrimSpace(req.UserCode)
	roleCode := strings.TrimSpace(req.RoleCode)
	if userCode == "" || roleCode == "" {
		return nil, errInvalid("user_code 与 role_code 均不能为空")
	}
	if err := l.svcCtx.Rbac.AssignUserRole(userCode, roleCode); err != nil {
		return nil, errInvalid(err.Error())
	}
	if err := svc.EnsureUserCasbinGroupings(l.svcCtx.Rbac, l.svcCtx.Casbin, userCode); err != nil {
		return nil, errInvalid("同步 Casbin 用户角色失败")
	}
	return &types.AssignIAMUserRoleReply{}, nil
}
