// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	orderdata "github.com/solikewind/happyeat/dal/model/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 列出订单
func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOrdersLogic) ListOrders(req *types.ListOrdersReq) (*types.ListOrdersReply, error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	current := int(req.Current)
	if current <= 0 {
		current = 1
	}
	offset := (current - 1) * pageSize

	filter := orderdata.ListOrdersFilter{
		Status:    req.Status,
		OrderType: req.OrderType,
		Offset:    offset,
		Limit:     pageSize,
	}
	if req.TableId > 0 {
		tid := int(req.TableId)
		filter.TableID = &tid
	}

	list, total, err := l.svcCtx.Order.List(l.ctx, filter)
	if err != nil {
		return nil, err
	}

	orders := make([]types.Order, 0, len(list))
	for _, e := range list {
		orders = append(orders, EntOrderToType(e))
	}

	return &types.ListOrdersReply{Orders: orders, Total: total}, nil
}
