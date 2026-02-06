// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menutype

import (
	"context"

	"happyeat/app/internal/svc"
	"happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMenuCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建菜单种类
func NewCreateMenuCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMenuCategoryLogic {
	return &CreateMenuCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMenuCategoryLogic) CreateMenuCategory(req *types.CreateMenuCategoryReq) (resp *types.CreateMenuCategoryReply, err error) {
	// todo: add your logic here and delete this line

	return
}
