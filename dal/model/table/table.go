// Package table 提供餐桌与餐桌分类的 data 逻辑。
package table

import (
	"context"

	"github.com/solikewind/happyeat/dal/model/ent"
	enttable "github.com/solikewind/happyeat/dal/model/ent/table"
	"github.com/solikewind/happyeat/dal/model/ent/tablecategory"
	"entgo.io/ent/dialect/sql"
)

// Table 餐桌数据访问。
type Table struct {
	c *ent.Client
}

// NewTable 创建 Table。
func NewTable(c *ent.Client) *Table {
	return &Table{c: c}
}

// CreateTableInput 创建餐桌入参。
type CreateTableInput struct {
	Code       string
	Status     string
	Capacity   int
	CategoryID int
	QRCode     string
}

// Create 创建餐桌。
func (t *Table) Create(ctx context.Context, in CreateTableInput) (*ent.Table, error) {
	create := t.c.Table.Create().
		SetCode(in.Code).
		SetCategoryID(in.CategoryID)
	if in.Status != "" {
		create = create.SetStatus(in.Status)
	} else {
		create = create.SetStatus("idle")
	}
	if in.Capacity > 0 {
		create = create.SetCapacity(in.Capacity)
	}
	if in.QRCode != "" {
		create = create.SetQrCode(in.QRCode)
	}

	return create.Save(ctx)
}

// GetByID 按 ID 获取餐桌（含 category）。
func (t *Table) GetByID(ctx context.Context, id int) (*ent.Table, error) {
	return t.c.Table.Query().
		Where(enttable.IDEQ(id)).
		WithCategory().
		Only(ctx)
}

// ListTablesFilter 列表筛选。
type ListTablesFilter struct {
	Code         string
	Status       string
	CategoryName string
	Offset       int
	Limit        int
}

// List 分页列出餐桌（含 category），返回列表与总数。
func (t *Table) List(ctx context.Context, f ListTablesFilter) ([]*ent.Table, int64, error) {
	q := t.c.Table.Query().WithCategory()
	if f.Code != "" {
		q = q.Where(enttable.CodeContains(f.Code))
	}
	if f.Status != "" {
		q = q.Where(enttable.StatusEQ(f.Status))
	}
	if f.CategoryName != "" {
		q = q.Where(enttable.HasCategoryWith(tablecategory.NameEQ(f.CategoryName)))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if f.Limit <= 0 {
		f.Limit = 10
	}

	list, err := q.Order(enttable.ByID(sql.OrderDesc())).Limit(f.Limit).Offset(f.Offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return list, int64(total), nil
}

// UpdateTableInput 更新餐桌入参。
type UpdateTableInput struct {
	Code       string
	Status     string
	Capacity   int
	CategoryID int
	QRCode     string
}

// Update 更新餐桌。
func (t *Table) Update(ctx context.Context, id int, in UpdateTableInput) error {
	upd := t.c.Table.UpdateOneID(id).
		SetCode(in.Code).
		SetStatus(in.Status).
		SetCapacity(in.Capacity).
		SetCategoryID(in.CategoryID)
	if in.QRCode != "" {
		upd = upd.SetQrCode(in.QRCode)
	} else {
		upd = upd.ClearQrCode()
	}

	_, err := upd.Save(ctx)
	return err
}

// Delete 删除餐桌。
func (t *Table) Delete(ctx context.Context, id int) error {
	return t.c.Table.DeleteOneID(id).Exec(ctx)
}
