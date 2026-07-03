// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListIAMPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 分页列出权限点（iam_permissions）
func NewListIAMPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListIAMPermissionsLogic {
	return &ListIAMPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListIAMPermissionsLogic) ListIAMPermissions(req *types.ListIAMPermissionsReq) (resp *types.ListIAMPermissionsReply, err error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	current := int(req.Current)
	if current <= 0 {
		current = 1
	}
	offset := (current - 1) * pageSize

	rows, total, err := l.svcCtx.Rbac.ListIAMPermissionsPage(l.ctx, offset, pageSize, req.Keyword)
	if err != nil {
		return nil, err
	}
	items := make([]types.PermissionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, types.PermissionItem{
			Id:          row.ID,
			Code:        row.Code,
			Description: row.Description,
		})
	}
	return &types.ListIAMPermissionsReply{
		Permissions: items,
		Total:       total,
	}, nil
}
