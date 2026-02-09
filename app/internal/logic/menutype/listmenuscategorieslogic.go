// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menutype

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMenusCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 列出菜单种类
func NewListMenusCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMenusCategoriesLogic {
	return &ListMenusCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMenusCategoriesLogic) ListMenusCategories(req *types.ListMenusCategoriesReq) (resp *types.ListMenusCategoriesReply, err error) {
	// todo: add your logic here and delete this line

	return
}
