// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	menudata "github.com/solikewind/happyeat/dal/model/menu"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMenusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 列出菜单
func NewListMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMenusLogic {
	return &ListMenusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMenusLogic) ListMenus(req *types.ListMenusReq) (*types.ListMenusReply, error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	current := int(req.Current)
	if current <= 0 {
		current = 1
	}
	offset := (current - 1) * pageSize

	list, total, err := l.svcCtx.Menu.List(l.ctx, menudata.ListMenusFilter{
		Name:         req.Name,
		CategoryName: req.Category,
		Offset:       offset,
		Limit:        pageSize,
	})
	if err != nil {
		return nil, err
	}

	menus := make([]types.Menu, 0, len(list))
	for _, e := range list {
		menus = append(menus, entMenuToType(e))
	}

	return &types.ListMenusReply{Menus: menus, Total: total}, nil
}
