// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListIAMUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 分页列出用户及其角色
func NewListIAMUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListIAMUsersLogic {
	return &ListIAMUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListIAMUsersLogic) ListIAMUsers(req *types.ListIAMUsersReq) (resp *types.ListIAMUsersReply, err error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	current := int(req.Current)
	if current <= 0 {
		current = 1
	}
	offset := (current - 1) * pageSize

	rows, total, err := l.svcCtx.Rbac.ListIAMUsersPage(l.ctx, offset, pageSize, req.Keyword)
	if err != nil {
		return nil, err
	}
	items := make([]types.IAMUserItem, 0, len(rows))
	for i := range rows {
		items = append(items, toUserItem(l.ctx, l.svcCtx, &rows[i]))
	}
	return &types.ListIAMUsersReply{
		Users: items,
		Total: total,
	}, nil
}
