// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

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

func (l *ListMenusLogic) ListMenus(req *types.ListMenusReq) (resp *types.ListMenusReply, err error) {
	// todo: add your logic here and delete this line

	return
}
