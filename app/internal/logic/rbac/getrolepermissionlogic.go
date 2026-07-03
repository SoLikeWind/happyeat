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

type GetRolePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取单个角色的权限列表
func NewGetRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRolePermissionLogic {
	return &GetRolePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRolePermissionLogic) GetRolePermission(req *types.GetRolePermissionReq) (resp *types.GetRolePermissionReply, err error) {
	role := strings.TrimSpace(req.Role)
	if role == "" {
		return nil, errInvalid("role 不能为空")
	}
	permissions, err := l.svcCtx.Rbac.GetRolePermissions(role)
	if err != nil {
		return nil, errInvalid(err.Error())
	}
	return &types.GetRolePermissionReply{
		RolePermission: types.RolePermission{
			Role:        role,
			Permissions: permissions,
		},
	}, nil
}
