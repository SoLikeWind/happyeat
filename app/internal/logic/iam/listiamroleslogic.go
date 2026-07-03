// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/pkg/casbinrules"
	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListIAMRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 分页列出角色（iam_roles）
func NewListIAMRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListIAMRolesLogic {
	return &ListIAMRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListIAMRolesLogic) ListIAMRoles(req *types.ListIAMRolesReq) (resp *types.ListIAMRolesReply, err error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	current := int(req.Current)
	if current <= 0 {
		current = 1
	}
	offset := (current - 1) * pageSize

	rows, total, err := l.svcCtx.Rbac.ListIAMRolesPage(l.ctx, offset, pageSize, req.Keyword)
	if err != nil {
		return nil, err
	}
	items := make([]types.IAMRoleItem, 0, len(rows))
	for _, row := range rows {
		if row.RoleCode == casbinrules.UnknownRoleCode {
			continue
		}
		items = append(items, types.IAMRoleItem{
			Id:       row.ID,
			RoleCode: row.RoleCode,
			RoleName: row.RoleName,
		})
	}
	return &types.ListIAMRolesReply{
		Roles: items,
		Total: total,
	}, nil
}
