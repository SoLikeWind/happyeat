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

type ResetRolePermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 重置角色权限（可单角色）
func NewResetRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetRolePermissionsLogic {
	return &ResetRolePermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetRolePermissionsLogic) ResetRolePermissions(req *types.ResetRolePermissionsReq) (resp *types.ResetRolePermissionsReply, err error) {
	role := strings.TrimSpace(req.Role)
	if err := l.svcCtx.Rbac.Reset(role); err != nil {
		return nil, errInvalid(err.Error())
	}
	if role == "" {
		roleMap, listErr := l.svcCtx.Rbac.List()
		if listErr != nil {
			return nil, errInvalid(listErr.Error())
		}
		for roleCode, permissions := range roleMap {
			if syncErr := svc.ReplaceRolePoliciesInCasbin(l.svcCtx.Casbin, roleCode, permissions); syncErr != nil {
				return nil, errInvalid("同步 Casbin 策略失败")
			}
		}
	} else {
		permissions, getErr := l.svcCtx.Rbac.GetRolePermissions(role)
		if getErr != nil {
			return nil, errInvalid(getErr.Error())
		}
		if syncErr := svc.ReplaceRolePoliciesInCasbin(l.svcCtx.Casbin, role, permissions); syncErr != nil {
			return nil, errInvalid("同步 Casbin 策略失败")
		}
	}
	if err := svc.EnsureActorCasbinGroupings(l.ctx, l.svcCtx.Rbac, l.svcCtx.Casbin); err != nil {
		return nil, errInvalid("同步 Casbin 用户角色失败")
	}
	return &types.ResetRolePermissionsReply{}, nil
}
