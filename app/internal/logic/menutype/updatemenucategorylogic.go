// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menutype

import (
	"context"

	"happyeat/app/internal/svc"
	"happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMenuCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新菜单种类
func NewUpdateMenuCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuCategoryLogic {
	return &UpdateMenuCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMenuCategoryLogic) UpdateMenuCategory(req *types.UpdateMenuCategoryReq) (resp *types.UpdateMenuCategoryReply, err error) {
	// todo: add your logic here and delete this line

	return
}
