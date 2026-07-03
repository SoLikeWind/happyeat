// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package rbac

import (
	"context"
	"sort"

	"github.com/solikewind/happyeat/app/internal/pkg/casbinrules"
	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRolePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取角色权限矩阵
func NewListRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolePermissionLogic {
	return &ListRolePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRolePermissionLogic) ListRolePermission() (resp *types.ListRolePermissionReply, err error) {
	roleMap, err := l.svcCtx.Rbac.List()
	if err != nil {
		return nil, errInvalid("查询角色权限失败")
	}
	roles := make([]string, 0, len(roleMap))
	for role := range roleMap {
		roles = append(roles, role)
	}
	sort.Strings(roles)

	result := make([]types.RolePermission, 0, len(roles))
	for _, role := range roles {
		if role == casbinrules.UnknownRoleCode {
			continue
		}
		result = append(result, types.RolePermission{
			Role:        role,
			Permissions: roleMap[role],
		})
	}

	return &types.ListRolePermissionReply{
		Roles: result,
	}, nil
}
