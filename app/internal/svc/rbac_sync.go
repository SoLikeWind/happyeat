package svc

import (
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
	grouping, err := enforcer.GetGroupingPolicy()
	if err != nil {
		return err
	}
	if len(grouping) > 0 {
		if _, err = enforcer.RemoveGroupingPolicies(grouping); err != nil {
			return err
		}
	}
	for userID, userRoles := range users {
		for _, role := range userRoles {
			if _, err = enforcer.AddGroupingPolicy(userID, role); err != nil {
				return err
			}
		}
	}
	return nil
}
