// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"
	"errors"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	"github.com/solikewind/happyeat/dal/model/ent"
	menudata "github.com/solikewind/happyeat/dal/model/menu"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMenuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建菜单
func NewCreateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMenuLogic {
	return &CreateMenuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMenuLogic) CreateMenu(req *types.CreateMenuReq) (*types.CreateMenuReply, error) {
	if req.Name == "" {
		return nil, errors.New("菜单名称不能为空")
	}
	if req.Price < 0 {
		return nil, errors.New("价格不能为负")
	}

	_, err := l.svcCtx.MenuType.GetByID(l.ctx, req.CategoryId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("分类不存在")
		}
		return nil, err
	}

	specs := make([]menudata.SpecInput, 0, len(req.Specs))
	for _, s := range req.Specs {
		specs = append(specs, menudata.SpecInput{SpecType: s.SpecType, SpecValue: s.SpecValue, PriceDelta: s.PriceDelta})
	}

	_, err = l.svcCtx.Menu.Create(l.ctx, menudata.CreateMenuInput{
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
		Price:       req.Price,
		CategoryID:  int(req.CategoryId),
		Specs:       specs,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateMenuReply{}, nil
}
