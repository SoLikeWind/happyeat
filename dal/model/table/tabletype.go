// Package table。tabletype.go 封装餐桌分类相关 data 逻辑。
package table

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/solikewind/happyeat/dal/model/ent"
	enttable "github.com/solikewind/happyeat/dal/model/ent/table"
	"github.com/solikewind/happyeat/dal/model/ent/tablecategory"
)

// TableType 餐桌分类数据访问。
type TableType struct {
	c *ent.Client
}

// NewTableType 创建 TableType。
func NewTableType(c *ent.Client) *TableType {
	return &TableType{c: c}
}

// CreateTableCategoryInput 创建分类入参。
type CreateTableCategoryInput struct {
	Name        string
	Description string
}

// Create 创建餐桌分类。
func (tt *TableType) Create(ctx context.Context, in CreateTableCategoryInput) (*ent.TableCategory, error) {
	create := tt.c.TableCategory.Create().SetName(in.Name)
	if in.Description != "" {
		create = create.SetDescription(in.Description)
	}

	return create.Save(ctx)
}

// GetByID 按 ID 获取分类。
func (tt *TableType) GetByID(ctx context.Context, id uint64) (*ent.TableCategory, error) {
	return tt.c.TableCategory.Get(ctx, int(id))
}

// ListTableCategoriesFilter 分类列表筛选。
type ListTableCategoriesFilter struct {
	Name   string
	Offset int64
	Limit  int64
}

// List 分页列出分类，返回列表与总数。
func (tt *TableType) List(ctx context.Context, f ListTableCategoriesFilter) ([]*ent.TableCategory, int64, error) {
	q := tt.c.TableCategory.Query()
	if f.Name != "" {
		q = q.Where(tablecategory.NameContains(f.Name))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if f.Limit <= 0 {
		f.Limit = 10
	}

	list, err := q.Order(tablecategory.ByID(sql.OrderDesc())).Limit(int(f.Limit)).Offset(int(f.Offset)).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return list, int64(total), nil
}

// Update 更新分类。
func (tt *TableType) Update(ctx context.Context, id int, name, description string) error {
	upd := tt.c.TableCategory.UpdateOneID(id).SetName(name)
	if description != "" {
		upd = upd.SetDescription(description)
	} else {
		upd = upd.ClearDescription()
	}

	_, err := upd.Save(ctx)
	return err
}

// Delete 删除分类。
func (tt *TableType) Delete(ctx context.Context, id int) error {
	return tt.c.TableCategory.DeleteOneID(id).Exec(ctx)
}

// CountTablesByCategoryID 统计某分类下餐桌数量。
func (tt *TableType) CountTablesByCategoryID(ctx context.Context, categoryID int) (int, error) {
	return tt.c.Table.Query().Where(enttable.HasCategoryWith(tablecategory.IDEQ(categoryID))).Count(ctx)
}
