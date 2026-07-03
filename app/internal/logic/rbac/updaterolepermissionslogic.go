// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package rbac

import (
	"context"
	"strings"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRolePermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新角色权限（全量覆盖）
func NewUpdateRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRolePermissionsLogic {
	return &UpdateRolePermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRolePermissionsLogic) UpdateRolePermissions(req *types.UpdateRolePermissionsReq) (resp *types.UpdateRolePermissionsReply, err error) {
	role := strings.TrimSpace(req.Role)
	if role == "" {
		return nil, errInvalid("role 不能为空")
	}
	if err := l.svcCtx.Rbac.UpdateRole(role, req.Permissions); err != nil {
		return nil, errInvalid(err.Error())
	}
	if err := svc.ReplaceRolePoliciesInCasbin(l.svcCtx.Casbin, role, req.Permissions); err != nil {
		return nil, errInvalid("同步 Casbin 策略失败")
	}
	if err := svc.EnsureActorCasbinGroupings(l.ctx, l.svcCtx.Rbac, l.svcCtx.Casbin); err != nil {
		return nil, errInvalid("同步 Casbin 用户角色失败")
	}
	return &types.UpdateRolePermissionsReply{}, nil
}
