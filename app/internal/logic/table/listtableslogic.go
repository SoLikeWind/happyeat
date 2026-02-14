// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package table

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	daltable "github.com/solikewind/happyeat/dal/model/table"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTablesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 列出餐桌
func NewListTablesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTablesLogic {
	return &ListTablesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTablesLogic) ListTables(req *types.ListTablesReq) (resp *types.ListTablesReply, err error) {
	current := req.Current
	pageSize := req.PageSize
	if current <= 0 {
		current = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (current - 1) * pageSize

	list, total, err := l.svcCtx.Table.List(l.ctx, daltable.ListTablesFilter{
		Code:         req.Code,
		Status:       req.Status,
		CategoryName: req.Category,
		Offset:       int(offset),
		Limit:        int(pageSize),
	})
	if err != nil {
		l.Errorf("ListTables Table.List err: %v", err)
		return nil, err
	}

	tables := make([]types.Table, 0, len(list))
	for _, e := range list {
		categoryID := uint64(0)
		if e.Edges.Category != nil {
			categoryID = uint64(e.Edges.Category.ID)
		}
		tables = append(tables, types.Table{
			Id:         uint64(e.ID),
			Code:       e.Code,
			Status:     e.Status,
			Capacity:   e.Capacity,
			CategoryId: categoryID,
			QrCode:     ptrToStr(e.QrCode),
			CreateAt:   e.CreatedAt.Unix(),
			UpdateAt:   e.UpdatedAt.Unix(),
		})
	}
	return &types.ListTablesReply{Tables: tables, Total: total}, nil
}

func ptrToStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
