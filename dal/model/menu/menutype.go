// Package menu。menutype.go 封装菜单分类相关 data 逻辑。
package menu

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/solikewind/happyeat/dal/model/ent"
	entmenu "github.com/solikewind/happyeat/dal/model/ent/menu"
	"github.com/solikewind/happyeat/dal/model/ent/menucategory"
)

// MenuType 菜单分类数据访问。
type MenuType struct {
	c *ent.Client
}

// NewMenuType 创建 MenuType。
func NewMenuType(c *ent.Client) *MenuType {
	return &MenuType{c: c}
}

// CreateMenuCategoryInput 创建分类入参。
type CreateMenuCategoryInput struct {
	Name        string
	Description string
}

// Create 创建菜单分类。
func (mt *MenuType) Create(ctx context.Context, in CreateMenuCategoryInput) (*ent.MenuCategory, error) {
	create := mt.c.MenuCategory.Create().SetName(in.Name)
	if in.Description != "" {
		create = create.SetDescription(in.Description)
	}

	return create.Save(ctx)
}

// GetByID 按 ID 获取分类。
func (mt *MenuType) GetByID(ctx context.Context, id uint64) (*ent.MenuCategory, error) {
	return mt.c.MenuCategory.Get(ctx, int(id))
}

// ListMenuCategoriesFilter 分类列表筛选。
type ListMenuCategoriesFilter struct {
	Name   string
	Offset int
	Limit  int
}

// List 分页列出分类，返回列表与总数。
func (mt *MenuType) List(ctx context.Context, f ListMenuCategoriesFilter) ([]*ent.MenuCategory, int64, error) {
	q := mt.c.MenuCategory.Query()
	if f.Name != "" {
		q = q.Where(menucategory.NameContains(f.Name))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if f.Limit <= 0 {
		f.Limit = 10
	}

	list, err := q.Order(menucategory.ByID(sql.OrderDesc())).Limit(f.Limit).Offset(f.Offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return list, int64(total), nil
}

// Update 更新分类。
func (mt *MenuType) Update(ctx context.Context, id int, name, description string) error {
	upd := mt.c.MenuCategory.UpdateOneID(id).SetName(name)
	if description != "" {
		upd = upd.SetDescription(description)
	} else {
		upd = upd.ClearDescription()
	}

	_, err := upd.Save(ctx)
	return err
}

// Delete 删除分类。
func (mt *MenuType) Delete(ctx context.Context, id int) error {
	return mt.c.MenuCategory.DeleteOneID(id).Exec(ctx)
}

// CountMenusByCategoryID 统计某分类下菜单数量。
func (mt *MenuType) CountMenusByCategoryID(ctx context.Context, categoryID int) (int, error) {
	return mt.c.Menu.Query().Where(entmenu.HasCategoryWith(menucategory.IDEQ(categoryID))).Count(ctx)
}
