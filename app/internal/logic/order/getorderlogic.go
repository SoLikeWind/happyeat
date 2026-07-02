// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	"github.com/solikewind/happyeat/common/util/timeutil"
	"github.com/solikewind/happyeat/dal/model/ent"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取单个订单
func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderLogic) GetOrder(req *types.GetOrderReq) (*types.GetOrderReply, error) {
	entOrder, err := l.svcCtx.Order.GetByID(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.GetOrderReply{Order: EntOrderToTypeForDisplay(l.ctx, l.svcCtx, entOrder)}, nil
}

// EntOrderToType 将 ent 订单转为 API 类型，供 order 与 workbench 等复用。
func EntOrderToType(e *ent.Order, dailySequence int) types.Order {
	out := types.Order{
		Id:            uint64(e.ID),
		OrderNo:       e.OrderNo,
		OrderType:     string(e.OrderType),
		Status:        string(e.Status),
		TotalAmount:   e.TotalAmount,
		ActualAmount:  e.ActualAmount,
		DailySequence: dailySequence,
		CreatedAt:     timeutil.TimeToString(e.CreatedAt),
		UpdatedAt:     timeutil.TimeToString(e.UpdatedAt),
	}

	if e.Remark != nil {
		out.Remark = *e.Remark
	}
	if e.SettlementID != nil && *e.SettlementID > 0 {
		out.SettlementId = *e.SettlementID
	}

	tbl, _ := e.Edges.TableOrErr()
	if tbl != nil {
		out.TableId = uint64(tbl.ID)
		out.TableCode = tbl.Code
		cat, _ := tbl.Edges.CategoryOrErr()
		if cat != nil {
			out.TableCategory = cat.Name
		}
	}

	items, _ := e.Edges.ItemsOrErr()
	for _, it := range items {
		oi := types.OrderItem{
			MenuName:  it.MenuName,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			Amount:    it.Amount,
		}
		if it.MenuID != nil && *it.MenuID > 0 {
			oi.MenuId = *it.MenuID
		}
		if it.SpecInfo != nil {
			oi.SpecInfo = *it.SpecInfo
		}
		out.Items = append(out.Items, oi)
	}

	return out
}
