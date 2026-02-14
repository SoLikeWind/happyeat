// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package table

import (
	"context"
	"errors"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	daltable "github.com/solikewind/happyeat/dal/model/table"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTableLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建餐桌
func NewCreateTableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTableLogic {
	return &CreateTableLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTableLogic) CreateTable(req *types.CreateTableReq) (resp *types.CreateTableReply, err error) {
	if req.Code == "" {
		return nil, errors.New("餐桌编号不能为空")
	}
	if req.Status == "" {
		return nil, errors.New("餐桌状态不能为空")
	}
	if req.Capacity <= 0 {
		return nil, errors.New("餐桌容量不能小于等于0")
	}
	_, err = l.svcCtx.Table.Create(l.ctx, daltable.CreateTableInput{
		Code:       req.Code,
		Status:     req.Status,
		Capacity:   int(req.Capacity),
		CategoryID: int(req.CategoryId),
		QRCode:     req.QrCode,
	})
	if err != nil {
		l.Errorf("CreateTable err: %v", err)
		return nil, err
	}
	return &types.CreateTableReply{}, nil
}
