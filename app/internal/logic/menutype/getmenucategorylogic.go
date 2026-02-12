// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menutype

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	"github.com/solikewind/happyeat/dal/model/ent"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取菜单种类
func NewGetMenuCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuCategoryLogic {
	return &GetMenuCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMenuCategoryLogic) GetMenuCategory(req *types.GetMenuCategoryReq) (*types.GetMenuCategoryReply, error) {
	cat, err := l.svcCtx.MenuType.GetByID(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.GetMenuCategoryReply{
		MenuCategory: entCategoryToType(cat),
	}, nil
}

func entCategoryToType(e *ent.MenuCategory) types.MenuCategory {
	out := types.MenuCategory{Id: uint64(e.ID), Name: e.Name}

	if e.Description != nil {
		out.Description = *e.Description
	}

	return out
}
