// Package casbinrules 定义权限码与 (obj,act) 映射，供 svc 与 routecheck 共用。
package casbinrules

// PolicyRule 单条 Casbin 风格策略（角色投影前的资源点）。
type PolicyRule struct {
	Obj string
	Act string
}

// ValidPermissions 允许写入 RBAC 的权限码集合。
var ValidPermissions = map[string]struct{}{
	"permission:view":      {},
	"audit:view":           {},
	"home:view":            {},
	"workbench:view":       {},
	"workbench:complete":   {},
	"orders:view":          {},
	"orders:create":        {},
	"orders:update":        {},
	"orders:update_status": {},
	"orders:print_kitchen": {},
	"order_desk:view":      {},
	"order_desk:create":    {},
	"menu:view":            {},
	"menu:edit":            {},
	"table:view":           {},
	"table:edit":           {},
	"spec:view":            {},
	"spec:edit":            {},
	"stats:view":           {},
	"settlements:view":     {},
	"settlements:edit":     {},
	"settlements:settle":   {},
}

// PermissionCatalog IAM 权限点种子（描述入库）。
type PermissionSpec struct {
	Code        string
	Description string
}

var PermissionCatalog = []PermissionSpec{
	{Code: "permission:view", Description: "查看和维护角色权限配置"},
	{Code: "audit:view", Description: "查看后台操作日志"},
	{Code: "home:view", Description: "查看首页"},
	{Code: "workbench:view", Description: "查看工作台订单"},
	{Code: "workbench:complete", Description: "完成工作台订单"},
	{Code: "orders:view", Description: "查看订单"},
	{Code: "orders:create", Description: "创建订单"},
	{Code: "orders:update", Description: "更新订单（追加菜单）"},
	{Code: "orders:update_status", Description: "更新订单状态"},
	{Code: "orders:print_kitchen", Description: "手动打印厨房小票（商鹏）"},
	{Code: "order_desk:view", Description: "点餐台查看"},
	{Code: "order_desk:create", Description: "点餐台创建订单"},
	{Code: "menu:view", Description: "查看菜单"},
	{Code: "menu:edit", Description: "编辑菜单"},
	{Code: "table:view", Description: "查看餐桌"},
	{Code: "table:edit", Description: "编辑餐桌"},
	{Code: "spec:view", Description: "查看规格模板"},
	{Code: "spec:edit", Description: "编辑规格模板"},
	{Code: "stats:view", Description: "查看经营统计"},
	{Code: "settlements:view", Description: "查看结账单"},
	{Code: "settlements:edit", Description: "创建结账单及管理订单"},
	{Code: "settlements:settle", Description: "结账单结账"},
}

// PermissionRules 权限码 -> HTTP 资源点（与 Casbin 中间件 obj/act 一致）。
var PermissionRules = map[string][]PolicyRule{
	"permission:view": {
		{Obj: "/central/v1/rbac/permission-catalog", Act: "GET"},
		{Obj: "/central/v1/rbac/role-permissions", Act: "GET"},
		{Obj: "/central/v1/rbac/role-permissions/:id", Act: "GET"},
		{Obj: "/central/v1/rbac/role-permissions/:id", Act: "PUT"},
		{Obj: "/central/v1/rbac/role-permissions/reset", Act: "POST"},
		{Obj: "/central/v1/rbac/casbin/sync", Act: "POST"},
		{Obj: "/central/v1/iam/permissions", Act: "GET"},
		{Obj: "/central/v1/iam/permissions/:id", Act: "GET"},
		{Obj: "/central/v1/iam/roles", Act: "GET"},
		{Obj: "/central/v1/iam/roles", Act: "POST"},
		{Obj: "/central/v1/iam/roles/:id", Act: "GET"},
		{Obj: "/central/v1/iam/roles/:id", Act: "PUT"},
		{Obj: "/central/v1/iam/roles/:id", Act: "DELETE"},
		{Obj: "/central/v1/iam/users", Act: "GET"},
		{Obj: "/central/v1/iam/users", Act: "POST"},
		{Obj: "/central/v1/iam/users/:id", Act: "GET"},
		{Obj: "/central/v1/iam/users/:id", Act: "PUT"},
		{Obj: "/central/v1/iam/users/:id", Act: "DELETE"},
		{Obj: "/central/v1/iam/user-roles", Act: "POST"},
		{Obj: "/central/v1/iam/user-roles", Act: "DELETE"},
		{Obj: "/central/v1/iam/users/:id/reset-password", Act: "POST"},
	},
	"audit:view": {{Obj: "/central/v1/operation-logs", Act: "GET"}},
	"workbench:view":       {{Obj: "/central/v1/workbench/orders", Act: "GET"}},
	"workbench:complete":   {{Obj: "/central/v1/order/:id/status", Act: "PUT"}},
	"orders:view":          {{Obj: "/central/v1/orders", Act: "GET"}, {Obj: "/central/v1/order/:id", Act: "GET"}},
	"orders:create":        {{Obj: "/central/v1/orders", Act: "POST"}},
	"orders:update":        {{Obj: "/central/v1/order/:id", Act: "PUT"}},
	"orders:update_status": {{Obj: "/central/v1/order/:id/status", Act: "PUT"}},
	"orders:print_kitchen": {{Obj: "/central/v1/order/:id/print", Act: "POST"}},
	"order_desk:view": {
		{Obj: "/central/v1/menu/categories", Act: "GET"},
		{Obj: "/central/v1/menus", Act: "GET"},
		{Obj: "/central/v1/tables", Act: "GET"},
	},
	"order_desk:create": {{Obj: "/central/v1/orders", Act: "POST"}},
	"menu:view": {
		{Obj: "/central/v1/menu/categories", Act: "GET"},
		{Obj: "/central/v1/menu/category/:id", Act: "GET"},
		{Obj: "/central/v1/menus", Act: "GET"},
		{Obj: "/central/v1/menu/:id", Act: "GET"},
		{Obj: "/central/v1/object/:id", Act: "GET"},
		{Obj: "/central/v1/object/:id/url", Act: "GET"},
	},
	"menu:edit": {
		{Obj: "/central/v1/menu/category", Act: "POST"},
		{Obj: "/central/v1/menu/category/:id", Act: "PUT"},
		{Obj: "/central/v1/menu/category/:id", Act: "DELETE"},
		{Obj: "/central/v1/menus", Act: "POST"},
		{Obj: "/central/v1/objects/upload", Act: "POST"},
		{Obj: "/central/v1/object/:id", Act: "DELETE"},
		{Obj: "/central/v1/menu/:id", Act: "PUT"},
		{Obj: "/central/v1/menu/:id", Act: "DELETE"},
	},
	"table:view": {
		{Obj: "/central/v1/table/categories", Act: "GET"},
		{Obj: "/central/v1/table/category/:id", Act: "GET"},
		{Obj: "/central/v1/tables", Act: "GET"},
		{Obj: "/central/v1/table/:id", Act: "GET"},
	},
	"table:edit": {
		{Obj: "/central/v1/table/category", Act: "POST"},
		{Obj: "/central/v1/table/category/:id", Act: "PUT"},
		{Obj: "/central/v1/table/category/:id", Act: "DELETE"},
		{Obj: "/central/v1/tables", Act: "POST"},
		{Obj: "/central/v1/table/:id", Act: "PUT"},
		{Obj: "/central/v1/table/:id", Act: "DELETE"},
	},
	"spec:view": {
		{Obj: "/central/v1/spec/category-spec", Act: "GET"},
		{Obj: "/central/v1/spec/category-spec/:id", Act: "GET"},
		{Obj: "/central/v1/spec/group/:id", Act: "GET"},
		{Obj: "/central/v1/spec/groups", Act: "GET"},
		{Obj: "/central/v1/spec/item/:id", Act: "GET"},
		{Obj: "/central/v1/spec/items", Act: "GET"},
	},
	"spec:edit": {
		{Obj: "/central/v1/spec/category-spec", Act: "POST"},
		{Obj: "/central/v1/spec/category-spec/:id", Act: "PUT"},
		{Obj: "/central/v1/spec/category-spec/:id", Act: "DELETE"},
		{Obj: "/central/v1/spec/group", Act: "POST"},
		{Obj: "/central/v1/spec/group/:id", Act: "PUT"},
		{Obj: "/central/v1/spec/group/:id", Act: "DELETE"},
		{Obj: "/central/v1/spec/item", Act: "POST"},
		{Obj: "/central/v1/spec/item/:id", Act: "PUT"},
		{Obj: "/central/v1/spec/item/:id", Act: "DELETE"},
	},
	"stats:view": {
		{Obj: "/central/v1/stats/daily", Act: "GET"},
		{Obj: "/central/v1/stats/daily/overview", Act: "GET"},
		{Obj: "/central/v1/stats/menus", Act: "GET"},
	},
	"settlements:view": {
		{Obj: "/central/v1/settlements", Act: "GET"},
		{Obj: "/central/v1/settlement/:id", Act: "GET"},
	},
	"settlements:edit": {
		{Obj: "/central/v1/settlements", Act: "POST"},
		{Obj: "/central/v1/settlement/:id", Act: "DELETE"},
		{Obj: "/central/v1/settlement/:id/orders", Act: "POST"},
		{Obj: "/central/v1/settlement/:id/orders/:id", Act: "DELETE"},
	},
	"settlements:settle": {
		{Obj: "/central/v1/settlement/:id/settle", Act: "POST"},
	},
}

// PresetRoleCodes 启动时种子写入的预置角色（不可删除）。
var PresetRoleCodes = map[string]struct{}{
	"super_admin": {},
	"manager":     {},
	"cashier":     {},
	"kitchen":     {},
	"waiter":      {},
}

// UnknownRoleCode 前端/鉴权兜底用的虚拟角色码，不应作为可分配 IAM 角色删除。
const UnknownRoleCode = "unknown"

// IsProtectedRole 系统保护角色（预置 + unknown），不可删除。
func IsProtectedRole(roleCode string) bool {
	if IsPresetRole(roleCode) {
		return true
	}
	return roleCode == UnknownRoleCode
}

// PresetRoleNames 预置角色的中文展示名（role_code 保持英文不变）。
var PresetRoleNames = map[string]string{
	"super_admin": "超级管理员",
	"manager":     "店长",
	"cashier":     "收银",
	"kitchen":     "后厨",
	"waiter":      "服务员",
}

// PresetRoleDisplayName 返回预置角色中文名；非预置或未知时回退 role_code。
func PresetRoleDisplayName(roleCode string) string {
	if name, ok := PresetRoleNames[roleCode]; ok && name != "" {
		return name
	}
	return roleCode
}

// IsPresetRole 是否为系统预置角色。
func IsPresetRole(roleCode string) bool {
	_, ok := PresetRoleCodes[roleCode]
	return ok
}

// CatalogEntry 权限目录项（含映射的 HTTP 资源点）。
type CatalogEntry struct {
	Code        string
	Description string
	Endpoints   []PolicyRule
}

// ListCatalog 返回可分配的权限目录（与 iam_permissions / PermissionRules 一致）。
func ListCatalog() []CatalogEntry {
	out := make([]CatalogEntry, 0, len(PermissionCatalog))
	for _, spec := range PermissionCatalog {
		rules := PermissionRules[spec.Code]
		eps := make([]PolicyRule, len(rules))
		copy(eps, rules)
		out = append(out, CatalogEntry{
			Code:        spec.Code,
			Description: spec.Description,
			Endpoints:   eps,
		})
	}
	return out
}

// AllPolicyKeys 返回所有 (canonicalObj, act) 用于与路由集合 diff。
func AllPolicyKeys(canonical func(obj, act string) string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, rules := range PermissionRules {
		for _, r := range rules {
			k := canonical(r.Obj, r.Act)
			out[k] = struct{}{}
		}
	}
	return out
}
