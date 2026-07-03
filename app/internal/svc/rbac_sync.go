package svc

import (
	"context"
	"fmt"
	"strings"

	"github.com/solikewind/happyeat/app/internal/pkg/casbinrules"
)

// SyncUserRoleGroupingAdd 在业务库已写入用户–角色后，仅向 Casbin 增加一条 g(user, role)。
// 与 SyncRolePoliciesToCasbin 全量相比：不触碰 p 策略，也不重写其他用户的 g。
func SyncUserRoleGroupingAdd(ce *CasbinEnforcer, userCode, roleCode string) error {
	_, err := ce.Enforcer.AddGroupingPolicy(userCode, roleCode)
	return err
}

// SyncUserRoleGroupingRemove 在业务库已解除绑定后，仅从 Casbin 移除对应 g(user, role)。
func SyncUserRoleGroupingRemove(ce *CasbinEnforcer, userCode, roleCode string) error {
	_, err := ce.Enforcer.RemoveGroupingPolicy(userCode, roleCode)
	return err
}

// EnsureUserCasbinGroupings 按 IAM 当前绑定对齐该用户在 Casbin 中的 g 策略：
// 补齐缺失的真实角色绑定，并移除 unknown 及 IAM 中已不存在的孤儿 g。
// 解决历史数据「Casbin 仅有 g(user,unknown)、IAM 已有 cashier/super_admin」时，
// 删掉 unknown 后真实角色未投影导致全站 403 的问题。
func EnsureUserCasbinGroupings(store *RbacStore, ce *CasbinEnforcer, userCode string) error {
	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return nil
	}
	roles, err := store.GetUserRoleCodes(userCode)
	if err != nil {
		return err
	}
	want := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		if role == casbinrules.UnknownRoleCode {
			continue
		}
		want[role] = struct{}{}
	}
	existing, err := ce.Enforcer.GetFilteredGroupingPolicy(0, userCode)
	if err != nil {
		return err
	}
	for _, g := range existing {
		if len(g) < 2 {
			continue
		}
		role := g[1]
		if _, ok := want[role]; !ok {
			if _, err := ce.Enforcer.RemoveGroupingPolicy(userCode, role); err != nil {
				return err
			}
		}
	}
	for role := range want {
		if _, err := ce.Enforcer.AddGroupingPolicy(userCode, role); err != nil {
			return err
		}
	}
	return nil
}

func userCodeFromContext(ctx context.Context) string {
	if v := ctx.Value("user_code"); v != nil {
		if s := stringifyContextValue(v); s != "" {
			return s
		}
	}
	if v := ctx.Value("sub"); v != nil {
		return stringifyContextValue(v)
	}
	return ""
}

func stringifyContextValue(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case fmt.Stringer:
		return strings.TrimSpace(x.String())
	default:
		s := strings.TrimSpace(fmt.Sprint(x))
		if s == "" || s == "<nil>" {
			return ""
		}
		return s
	}
}

// EnsureActorCasbinGroupings 按 IAM 对齐当前登录用户在 Casbin 中的 g 策略（角色权限变更后防止操作者自身 403）。
func EnsureActorCasbinGroupings(ctx context.Context, store *RbacStore, ce *CasbinEnforcer) error {
	if userCode := userCodeFromContext(ctx); userCode != "" {
		return EnsureUserCasbinGroupings(store, ce, userCode)
	}
	return nil
}

// RemoveAllUserCasbinGroupings 移除某用户在 Casbin 中的全部分组（删用户时用，避免全量重建）。
func RemoveAllUserCasbinGroupings(ce *CasbinEnforcer, userCode string) error {
	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return nil
	}
	existing, err := ce.Enforcer.GetFilteredGroupingPolicy(0, userCode)
	if err != nil {
		return err
	}
	for _, g := range existing {
		if len(g) < 2 {
			continue
		}
		if _, err := ce.Enforcer.RemoveGroupingPolicy(userCode, g[1]); err != nil {
			return err
		}
	}
	return nil
}

// EnsureUserCasbinGroupingsForUsers 批量对齐多个用户的 Casbin g 策略（去重、忽略空 user_code）。
func EnsureUserCasbinGroupingsForUsers(store *RbacStore, ce *CasbinEnforcer, userCodes []string) error {
	seen := make(map[string]struct{}, len(userCodes))
	for _, userCode := range userCodes {
		userCode = strings.TrimSpace(userCode)
		if userCode == "" {
			continue
		}
		if _, ok := seen[userCode]; ok {
			continue
		}
		seen[userCode] = struct{}{}
		if err := EnsureUserCasbinGroupings(store, ce, userCode); err != nil {
			return err
		}
	}
	return nil
}

// ReplaceRolePoliciesInCasbin 仅替换某角色在 Casbin 中的 p 策略，不触碰用户 g 绑定。
func ReplaceRolePoliciesInCasbin(ce *CasbinEnforcer, roleCode string, permissions []string) error {
	enforcer := ce.Enforcer
	filtered, err := enforcer.GetFilteredPolicy(0, roleCode)
	if err != nil {
		return err
	}
	if len(filtered) > 0 {
		if _, err = enforcer.RemovePolicies(filtered); err != nil {
			return err
		}
	}
	for _, policy := range BuildPoliciesForPermissions(permissions) {
		if _, err = enforcer.AddPolicy(roleCode, policy.Obj, policy.Act); err != nil {
			return err
		}
	}
	return nil
}

// RemoveRolePoliciesFromCasbin 增量移除某角色的 p 策略与相关用户的 g 绑定（删除角色时用，避免全量重建 Casbin）。
func RemoveRolePoliciesFromCasbin(ce *CasbinEnforcer, roleCode string, userCodes, permissions []string) error {
	for _, policy := range BuildPoliciesForPermissions(permissions) {
		if _, err := ce.Enforcer.RemovePolicy(roleCode, policy.Obj, policy.Act); err != nil {
			return err
		}
	}
	for _, userCode := range userCodes {
		if _, err := ce.Enforcer.RemoveGroupingPolicy(userCode, roleCode); err != nil {
			return err
		}
	}
	return nil
}

func SyncRolePoliciesToCasbin(store *RbacStore, ce *CasbinEnforcer) error {
	enforcer := ce.Enforcer
	roles, err := store.List()
	if err != nil {
		return err
	}
	// Casbin v2.135+：RemoveFilteredPolicy 必须带 fieldValues，不能再用 (0) 表示「删全部」。
	policies, err := enforcer.GetPolicy()
	if err != nil {
		return err
	}
	if len(policies) > 0 {
		if _, err = enforcer.RemovePolicies(policies); err != nil {
			return err
		}
	}
	for role, permissions := range roles {
		policies := BuildPoliciesForPermissions(permissions)
		for _, policy := range policies {
			if _, err = enforcer.AddPolicy(role, policy.Obj, policy.Act); err != nil {
				return err
			}
		}
	}
	users, err := store.ListUserRoles()
	if err != nil {
		return err
	}
	for userCode := range users {
		if err := EnsureUserCasbinGroupings(store, ce, userCode); err != nil {
			return err
		}
	}
	return nil
}
