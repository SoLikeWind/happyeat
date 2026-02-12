// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package workbench

import (
	"context"
	"strings"

	"github.com/solikewind/happyeat/app/internal/logic/order"
	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	orderdata "github.com/solikewind/happyeat/dal/model/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWorkbenchOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 工作台订单列表（默认待处理：created/paid/preparing）；出单用 更新订单状态 置为 completed
func NewListWorkbenchOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWorkbenchOrdersLogic {
	return &ListWorkbenchOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 工作台默认展示的状态
var workbenchDefaultStatuses = []string{"created", "paid", "preparing"}

func (l *ListWorkbenchOrdersLogic) ListWorkbenchOrders(req *types.ListWorkbenchOrdersReq) (*types.ListWorkbenchOrdersReply, error) {
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
		Offset: offset,
		Limit:  pageSize,
	}
	if strings.TrimSpace(req.Status) != "" {
		filter.Status = strings.TrimSpace(req.Status)
	} else {
		filter.Statuses = workbenchDefaultStatuses
	}

	list, total, err := l.svcCtx.Order.List(l.ctx, filter)
	if err != nil {
		return nil, err
	}

	orders := make([]types.Order, 0, len(list))
	for _, e := range list {
		orders = append(orders, order.EntOrderToType(e))
	}

	return &types.ListWorkbenchOrdersReply{Orders: orders, Total: total}, nil
}
