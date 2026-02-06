// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menutype

import (
	"context"

	"happyeat/app/internal/svc"
	"happyeat/app/internal/types"

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

func (l *GetMenuCategoryLogic) GetMenuCategory(req *types.GetMenuCategoryReq) (resp *types.GetMenuCategoryReply, err error) {
	// todo: add your logic here and delete this line

	return
}
