// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tablecategory

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	"github.com/solikewind/happyeat/dal/model/ent"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTableCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取餐桌类别
func NewGetTableCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTableCategoryLogic {
	return &GetTableCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTableCategoryLogic) GetTableCategory(req *types.GetTableCategoryReq) (resp *types.GetTableCategoryReply, err error) {
	category, err := l.svcCtx.TableType.GetByID(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.GetTableCategoryReply{
		TableCategory: entTableCategoryToType(category),
	}, nil
}

func entTableCategoryToType(e *ent.TableCategory) types.TableCategory {
	out := types.TableCategory{Id: uint64(e.ID), Name: e.Name}

	if e.Description != nil {
		out.Description = *e.Description
	}

	return out
}
