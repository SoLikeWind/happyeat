// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tablecategory

import (
	"context"
	"errors"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	tabledata "github.com/solikewind/happyeat/dal/model/table"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTableCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建餐桌类别
func NewCreateTableCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTableCategoryLogic {
	return &CreateTableCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTableCategoryLogic) CreateTableCategory(req *types.CreateTableCategoryReq) (*types.CreateTableCategoryReply, error) {
	if req.Name == "" {
		return nil, errors.New("分类名称不能为空")
	}

	_, err := l.svcCtx.TableType.Create(l.ctx, tabledata.CreateTableCategoryInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateTableCategoryReply{}, nil
}
