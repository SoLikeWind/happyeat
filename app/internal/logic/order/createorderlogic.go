// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"
	"errors"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	orderdata "github.com/solikewind/happyeat/dal/model/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建订单
func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (*types.CreateOrderReply, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("订单至少需要一项")
	}
	if req.OrderType != "dine_in" && req.OrderType != "takeaway" {
		return nil, errors.New("order_type 应为 dine_in 或 takeaway")
	}

	items := make([]orderdata.ItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, orderdata.ItemInput{
			MenuName:  it.MenuName,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			SpecInfo:  it.SpecInfo,
		})
	}

	var tableID *int
	if req.TableId > 0 {
		tid := int(req.TableId)
		tableID = &tid
	}

	entOrder, err := l.svcCtx.Order.Create(l.ctx, orderdata.CreateOrderInput{
		OrderType:   req.OrderType,
		TableID:     tableID,
		Items:       items,
		TotalAmount: req.TotalAmount,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateOrderReply{Order: EntOrderToType(entOrder)}, nil
}
