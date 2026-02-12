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

type UpdateMenuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新菜单
func NewUpdateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuLogic {
	return &UpdateMenuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMenuLogic) UpdateMenu(req *types.UpdateMenuReq) (*types.UpdateMenuReply, error) {
	m := req.Menu

	if m.Name == "" {
		return nil, errors.New("菜单名称不能为空")
	}
	if m.Price < 0 {
		return nil, errors.New("价格不能为负")
	}

	_, err := l.svcCtx.MenuType.GetByID(l.ctx, m.CategoryId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("分类不存在")
		}
		return nil, err
	}

	specs := make([]menudata.SpecInput, 0, len(m.Specs))
	for _, s := range m.Specs {
		specs = append(specs, menudata.SpecInput{SpecType: s.SpecType, SpecValue: s.SpecValue, PriceDelta: s.PriceDelta})
	}

	err = l.svcCtx.Menu.Update(l.ctx, int(req.Id), menudata.UpdateMenuInput{
		Name:        m.Name,
		Description: m.Description,
		Image:       m.Image,
		Price:       m.Price,
		CategoryID:  int(m.CategoryId),
		Specs:       specs,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, err
		}
		return nil, err
	}

	return &types.UpdateMenuReply{}, nil
}
