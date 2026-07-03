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

type CreateIAMRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建角色（无权限点，需再通过 RBAC 矩阵配置）
func NewCreateIAMRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateIAMRoleLogic {
	return &CreateIAMRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateIAMRoleLogic) CreateIAMRole(req *types.CreateIAMRoleReq) (resp *types.CreateIAMRoleReply, err error) {
	roleCode := strings.TrimSpace(req.RoleCode)
	roleName := strings.TrimSpace(req.RoleName)
	id, err := l.svcCtx.Rbac.CreateRole(roleCode, roleName)
	if err != nil {
		return nil, errInvalid(err.Error())
	}
	roleCode = strings.ToLower(roleCode)
	if len(req.Permissions) > 0 {
		if err := l.svcCtx.Rbac.UpdateRole(roleCode, req.Permissions); err != nil {
			return nil, errInvalid(err.Error())
		}
		if err := svc.ReplaceRolePoliciesInCasbin(l.svcCtx.Casbin, roleCode, req.Permissions); err != nil {
			return nil, errInvalid("同步 Casbin 策略失败")
		}
	}
	if err := svc.EnsureActorCasbinGroupings(l.ctx, l.svcCtx.Rbac, l.svcCtx.Casbin); err != nil {
		return nil, errInvalid("同步 Casbin 用户角色失败")
	}
	return &types.CreateIAMRoleReply{Id: id}, nil
}
