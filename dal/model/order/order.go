// Package order 提供订单与订单明细的 data 逻辑。
package order

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/solikewind/happyeat/dal/model/ent"
	entorder "github.com/solikewind/happyeat/dal/model/ent/order"
	enttable "github.com/solikewind/happyeat/dal/model/ent/table"
	"entgo.io/ent/dialect/sql"
)

// Order 订单数据访问。
type Order struct {
	c *ent.Client
}

// NewOrder 创建 Order。
func NewOrder(c *ent.Client) *Order {
	return &Order{c: c}
}

// ItemInput 订单项入参（下单时从菜单快照）
type ItemInput struct {
	MenuName  string
	Quantity  int
	UnitPrice float64
	SpecInfo  string
}

// CreateOrderInput 创建订单入参。
type CreateOrderInput struct {
	OrderType  string       // dine_in | takeaway
	TableID    *int         // 堂食时必填，外带为 nil
	Items      []ItemInput  // 至少一项
	TotalAmount float64
	Remark     string
}

// genOrderNo 生成订单号（简单示例：ORD+毫秒时间戳+3位随机）
func genOrderNo() string {
	return fmt.Sprintf("ORD%d%03d", time.Now().UnixMilli(), rand.Intn(1000))
}

// Create 创建订单及明细（事务内）。业务规则（order_type、堂食必填桌台等）由调用方 Logic 校验。
func (o *Order) Create(ctx context.Context, in CreateOrderInput) (*ent.Order, error) {
	if len(in.Items) == 0 {
		return nil, fmt.Errorf("items required")
	}

	tx, err := o.c.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	create := tx.Order.Create().
		SetOrderNo(genOrderNo()).
		SetOrderType(in.OrderType).
		SetStatus("created").
		SetTotalAmount(in.TotalAmount)
	if in.TableID != nil && *in.TableID > 0 {
		create = create.SetTableID(*in.TableID)
	}
	if in.Remark != "" {
		create = create.SetRemark(in.Remark)
	}

	entOrder, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	for i, item := range in.Items {
		itemCreate := tx.OrderItem.Create().
			SetOrderID(entOrder.ID).
			SetMenuName(item.MenuName).
			SetQuantity(item.Quantity).
			SetUnitPrice(item.UnitPrice).
			SetAmount(item.UnitPrice * float64(item.Quantity)).
			SetSort(i)
		if item.SpecInfo != "" {
			itemCreate = itemCreate.SetSpecInfo(item.SpecInfo)
		}
		if _, err = itemCreate.Save(ctx); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return o.c.Order.Query().Where(entorder.IDEQ(entOrder.ID)).WithTable().WithItems().Only(ctx)
}

// GetByID 按 ID 获取订单（含 table、items）。
func (o *Order) GetByID(ctx context.Context, id int) (*ent.Order, error) {
	return o.c.Order.Query().
		Where(entorder.IDEQ(id)).
		WithTable().
		WithItems().
		Only(ctx)
}

// ListOrdersFilter 列表筛选。
type ListOrdersFilter struct {
	Status    string   // 单状态
	Statuses  []string // 多状态（与 Status 二选一，Statuses 优先）
	OrderType string
	TableID   *int
	Offset    int
	Limit     int
}

// List 分页列出订单（含 table、items），返回列表与总数。
func (o *Order) List(ctx context.Context, f ListOrdersFilter) ([]*ent.Order, int64, error) {
	q := o.c.Order.Query().WithTable().WithItems()
	if len(f.Statuses) > 0 {
		q = q.Where(entorder.StatusIn(f.Statuses...))
	} else if f.Status != "" {
		q = q.Where(entorder.StatusEQ(f.Status))
	}
	if f.OrderType != "" {
		q = q.Where(entorder.OrderTypeEQ(f.OrderType))
	}
	if f.TableID != nil && *f.TableID > 0 {
		q = q.Where(entorder.HasTableWith(enttable.IDEQ(*f.TableID)))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if f.Limit <= 0 {
		f.Limit = 10
	}

	list, err := q.Order(entorder.ByID(sql.OrderDesc())).Limit(f.Limit).Offset(f.Offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return list, int64(total), nil
}

// UpdateStatus 更新订单状态。
func (o *Order) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := o.c.Order.UpdateOneID(id).SetStatus(status).Save(ctx)
	return err
}
