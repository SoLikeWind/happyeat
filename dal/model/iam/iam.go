// Package iam 提供 IAM 用户/角色/权限的数据访问（对 ent 的薄封装）。
// 业务校验、Casbin 投影/种子编排等服务逻辑在 app/internal/svc/rbac_store.go。
package iam

import (
	"context"
	"errors"

	"github.com/solikewind/happyeat/dal/model/ent"
	"github.com/solikewind/happyeat/dal/model/ent/iampermission"
	"github.com/solikewind/happyeat/dal/model/ent/iamrole"
	"github.com/solikewind/happyeat/dal/model/ent/iamuser"
)

// ErrPermissionsNotFound 覆盖角色权限时存在无效权限码。
var ErrPermissionsNotFound = errors.New("permissions not found")

// IAM IAM 数据访问层。
type IAM struct {
	c *ent.Client
}

// NewIAM 构造 IAM DAL。
func NewIAM(c *ent.Client) *IAM {
	return &IAM{c: c}
}

// CreateRow 创建用户入参（密码为已哈希值，手机号为空表示不设置）。
type CreateRow struct {
	UserCode       string
	Phone          string
	DisplayName    string
	PasswordHash   string
	AvatarObjectID uint64
	AvatarURL      string
}

// UpdateRow 更新用户入参（nil 表示该字段不改）。
type UpdateRow struct {
	DisplayName    *string
	Phone          *string
	AvatarObjectID *uint64
	AvatarURL      *string
}

// ---------------- 角色 ----------------

// RoleExists 角色是否存在。
func (m *IAM) RoleExists(ctx context.Context, roleCode string) (bool, error) {
	return m.c.IAMRole.Query().Where(iamrole.RoleCodeEQ(roleCode)).Exist(ctx)
}

// GetRoleByID 按 ID 获取角色。
func (m *IAM) GetRoleByID(ctx context.Context, id uint64) (*ent.IAMRole, error) {
	return m.c.IAMRole.Query().Where(iamrole.IDEQ(id)).Only(ctx)
}

// UpdateRoleNameByID 按 ID 更新角色展示名。
func (m *IAM) UpdateRoleNameByID(ctx context.Context, id uint64, roleName string) error {
	_, err := m.c.IAMRole.UpdateOneID(id).SetRoleName(roleName).Save(ctx)
	return err
}

// CreateRole 创建角色。
func (m *IAM) CreateRole(ctx context.Context, roleCode, roleName string) (*ent.IAMRole, error) {
	return m.c.IAMRole.Create().SetRoleCode(roleCode).SetRoleName(roleName).Save(ctx)
}

// EnsureRole 角色不存在时创建。
func (m *IAM) EnsureRole(ctx context.Context, roleCode, roleName string) error {
	exists, err := m.RoleExists(ctx, roleCode)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = m.CreateRole(ctx, roleCode, roleName)
	return err
}

// DeleteRoleByID 软删角色。
func (m *IAM) DeleteRoleByID(ctx context.Context, id uint64) error {
	return m.c.IAMRole.DeleteOneID(id).Exec(ctx)
}

// ListRolesPage 分页角色（按 role_code 升序），keyword 模糊匹配 code/name。
func (m *IAM) ListRolesPage(ctx context.Context, offset, limit int, keyword string) ([]*ent.IAMRole, int, error) {
	q := m.c.IAMRole.Query()
	if keyword != "" {
		q = q.Where(iamrole.Or(
			iamrole.RoleCodeContainsFold(keyword),
			iamrole.RoleNameContainsFold(keyword),
		))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Asc(iamrole.FieldRoleCode)).Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// RolePermissionsMap 角色 -> 权限码集合（升序）。
func (m *IAM) RolePermissionsMap(ctx context.Context) (map[string][]string, error) {
	roles, err := m.c.IAMRole.Query().
		WithPermissions(func(q *ent.IAMPermissionQuery) {
			q.Order(ent.Asc(iampermission.FieldPermissionCode))
		}).
		Order(ent.Asc(iamrole.FieldRoleCode)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(roles))
	for _, role := range roles {
		out[role.RoleCode] = make([]string, 0, len(role.Edges.Permissions))
		for _, p := range role.Edges.Permissions {
			out[role.RoleCode] = append(out[role.RoleCode], p.PermissionCode)
		}
	}
	return out, nil
}

// RolePermissions 按角色编码获取权限码集合（升序）。
func (m *IAM) RolePermissions(ctx context.Context, roleCode string) ([]string, error) {
	role, err := m.c.IAMRole.Query().
		Where(iamrole.RoleCodeEQ(roleCode)).
		WithPermissions(func(q *ent.IAMPermissionQuery) {
			q.Order(ent.Asc(iampermission.FieldPermissionCode))
		}).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(role.Edges.Permissions))
	for _, p := range role.Edges.Permissions {
		out = append(out, p.PermissionCode)
	}
	return out, nil
}

// RolePermissionsCount 角色-权限关联总数（用于判断是否需要种子）。
func (m *IAM) RolePermissionsCount(ctx context.Context) (int, error) {
	return m.c.IAMRole.Query().QueryPermissions().Count(ctx)
}

// SetRolePermissions 全量覆盖角色权限（permCodes 需已去重且合法）。
func (m *IAM) SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error {
	roleEnt, err := m.c.IAMRole.Query().Where(iamrole.RoleCodeEQ(roleCode)).Only(ctx)
	if err != nil {
		return err
	}
	tx, err := m.c.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updater := tx.IAMRole.UpdateOneID(roleEnt.ID).ClearPermissions()
	if len(permCodes) > 0 {
		perms, err := tx.IAMPermission.Query().Where(iampermission.PermissionCodeIn(permCodes...)).All(ctx)
		if err != nil {
			return err
		}
		if len(perms) != len(permCodes) {
			return ErrPermissionsNotFound
		}
		ids := make([]uint64, 0, len(perms))
		for _, p := range perms {
			ids = append(ids, p.ID)
		}
		updater = updater.AddPermissionIDs(ids...)
	}
	if _, err = updater.Save(ctx); err != nil {
		return err
	}
	return tx.Commit()
}

// ---------------- 权限点 ----------------

// GetPermissionByID 按 ID 获取权限点。
func (m *IAM) GetPermissionByID(ctx context.Context, id uint64) (*ent.IAMPermission, error) {
	return m.c.IAMPermission.Query().Where(iampermission.IDEQ(id)).Only(ctx)
}

// EnsurePermission 权限点不存在时创建。
func (m *IAM) EnsurePermission(ctx context.Context, code, description string) error {
	exists, err := m.c.IAMPermission.Query().Where(iampermission.PermissionCodeEQ(code)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = m.c.IAMPermission.Create().SetPermissionCode(code).SetDescription(description).Save(ctx)
	return err
}

// ListPermissionsPage 分页权限点（按 code 升序），keyword 模糊匹配 code/description。
func (m *IAM) ListPermissionsPage(ctx context.Context, offset, limit int, keyword string) ([]*ent.IAMPermission, int, error) {
	q := m.c.IAMPermission.Query()
	if keyword != "" {
		q = q.Where(iampermission.Or(
			iampermission.PermissionCodeContainsFold(keyword),
			iampermission.DescriptionContainsFold(keyword),
		))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Asc(iampermission.FieldPermissionCode)).Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ---------------- 用户 ----------------

// UserExistsByCode user_code 是否存在。
func (m *IAM) UserExistsByCode(ctx context.Context, userCode string) (bool, error) {
	return m.c.IAMUser.Query().Where(iamuser.UserCodeEQ(userCode)).Exist(ctx)
}

// UserExistsByID 用户 ID 是否存在。
func (m *IAM) UserExistsByID(ctx context.Context, id uint64) (bool, error) {
	return m.c.IAMUser.Query().Where(iamuser.IDEQ(id)).Exist(ctx)
}

// PhoneExists 手机号是否已被占用。
func (m *IAM) PhoneExists(ctx context.Context, phone string) (bool, error) {
	return m.c.IAMUser.Query().Where(iamuser.PhoneEQ(phone)).Exist(ctx)
}

// PhoneExistsExceptID 手机号是否被除指定 ID 外的用户占用。
func (m *IAM) PhoneExistsExceptID(ctx context.Context, phone string, id uint64) (bool, error) {
	return m.c.IAMUser.Query().Where(iamuser.PhoneEQ(phone), iamuser.IDNEQ(id)).Exist(ctx)
}

func userQueryWithRoles(q *ent.IAMUserQuery, withRoles bool) *ent.IAMUserQuery {
	if withRoles {
		q = q.WithRoles(func(rq *ent.IAMRoleQuery) { rq.Order(ent.Asc(iamrole.FieldRoleCode)) })
	}
	return q
}

// GetByCode 按 user_code 获取用户（可选带角色）。
func (m *IAM) GetByCode(ctx context.Context, userCode string, withRoles bool) (*ent.IAMUser, error) {
	return userQueryWithRoles(m.c.IAMUser.Query().Where(iamuser.UserCodeEQ(userCode)), withRoles).Only(ctx)
}

// GetByPhone 按手机号获取用户（可选带角色）。
func (m *IAM) GetByPhone(ctx context.Context, phone string, withRoles bool) (*ent.IAMUser, error) {
	return userQueryWithRoles(m.c.IAMUser.Query().Where(iamuser.PhoneEQ(phone)), withRoles).Only(ctx)
}

// GetByID 按 ID 获取用户（可选带角色）。
func (m *IAM) GetByID(ctx context.Context, id uint64, withRoles bool) (*ent.IAMUser, error) {
	return userQueryWithRoles(m.c.IAMUser.Query().Where(iamuser.IDEQ(id)), withRoles).Only(ctx)
}

// Create 创建用户主体。
func (m *IAM) Create(ctx context.Context, in CreateRow) (*ent.IAMUser, error) {
	create := m.c.IAMUser.Create().
		SetUserCode(in.UserCode).
		SetDisplayName(in.DisplayName).
		SetAvatarObjectID(in.AvatarObjectID).
		SetAvatarURL(in.AvatarURL)
	if in.Phone != "" {
		create = create.SetPhone(in.Phone)
	}
	if in.PasswordHash != "" {
		create = create.SetPasswordHash(in.PasswordHash)
	}
	return create.Save(ctx)
}

// UpdatePasswordByCode 按 user_code 更新密码哈希。
func (m *IAM) UpdatePasswordByCode(ctx context.Context, userCode, hash string) error {
	row, err := m.c.IAMUser.Query().Where(iamuser.UserCodeEQ(userCode)).Only(ctx)
	if err != nil {
		return err
	}
	_, err = m.c.IAMUser.UpdateOneID(row.ID).SetPasswordHash(hash).Save(ctx)
	return err
}

// UpdatePasswordByID 按 ID 更新密码哈希。
func (m *IAM) UpdatePasswordByID(ctx context.Context, id uint64, hash string) error {
	_, err := m.c.IAMUser.UpdateOneID(id).SetPasswordHash(hash).Save(ctx)
	return err
}

// UpdateAvatarByCode 按 user_code 更新头像对象与 URL。
func (m *IAM) UpdateAvatarByCode(ctx context.Context, userCode string, objectID uint64, url string) error {
	row, err := m.c.IAMUser.Query().Where(iamuser.UserCodeEQ(userCode)).Only(ctx)
	if err != nil {
		return err
	}
	_, err = m.c.IAMUser.UpdateOneID(row.ID).SetAvatarObjectID(objectID).SetAvatarURL(url).Save(ctx)
	return err
}

// UpdateByID 按 ID 更新用户字段（nil 表示不改）。
func (m *IAM) UpdateByID(ctx context.Context, id uint64, in UpdateRow) error {
	upd := m.c.IAMUser.UpdateOneID(id)
	if in.DisplayName != nil {
		upd = upd.SetDisplayName(*in.DisplayName)
	}
	if in.Phone != nil {
		upd = upd.SetPhone(*in.Phone)
	}
	if in.AvatarObjectID != nil {
		upd = upd.SetAvatarObjectID(*in.AvatarObjectID)
	}
	if in.AvatarURL != nil {
		upd = upd.SetAvatarURL(*in.AvatarURL)
	}
	_, err := upd.Save(ctx)
	return err
}

// DeleteByID 清除角色关联后软删用户。
func (m *IAM) DeleteByID(ctx context.Context, id uint64) error {
	if _, err := m.c.IAMUser.UpdateOneID(id).ClearRoles().Save(ctx); err != nil {
		return err
	}
	return m.c.IAMUser.DeleteOneID(id).Exec(ctx)
}

// ListUsersPage 分页用户（按 user_code 升序，带角色），keyword 模糊匹配 user_code/display_name。
func (m *IAM) ListUsersPage(ctx context.Context, offset, limit int, keyword string) ([]*ent.IAMUser, int, error) {
	q := m.c.IAMUser.Query()
	if keyword != "" {
		q = q.Where(iamuser.Or(
			iamuser.UserCodeContainsFold(keyword),
			iamuser.DisplayNameContainsFold(keyword),
			iamuser.PhoneContainsFold(keyword),
		))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Asc(iamuser.FieldUserCode)).Offset(offset).Limit(limit).
		WithRoles(func(rq *ent.IAMRoleQuery) { rq.Order(ent.Asc(iamrole.FieldRoleCode)) }).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// UserRolesMap 用户 -> 角色码集合（升序）。
func (m *IAM) UserRolesMap(ctx context.Context) (map[string][]string, error) {
	users, err := m.c.IAMUser.Query().
		WithRoles(func(q *ent.IAMRoleQuery) { q.Order(ent.Asc(iamrole.FieldRoleCode)) }).
		Order(ent.Asc(iamuser.FieldUserCode)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(users))
	for _, u := range users {
		out[u.UserCode] = make([]string, 0, len(u.Edges.Roles))
		for _, r := range u.Edges.Roles {
			out[u.UserCode] = append(out[u.UserCode], r.RoleCode)
		}
	}
	return out, nil
}

// ---------------- 用户-角色关联 ----------------

// UserHasRole 用户是否已绑定该角色。
func (m *IAM) UserHasRole(ctx context.Context, userCode, roleCode string) (bool, error) {
	return m.c.IAMUser.Query().
		Where(iamuser.UserCodeEQ(userCode), iamuser.HasRolesWith(iamrole.RoleCodeEQ(roleCode))).
		Exist(ctx)
}

// AddUserRole 绑定角色（调用前需确保用户与角色均存在）。
func (m *IAM) AddUserRole(ctx context.Context, userCode, roleCode string) error {
	userEnt, err := m.c.IAMUser.Query().Where(iamuser.UserCodeEQ(userCode)).Only(ctx)
	if err != nil {
		return err
	}
	roleEnt, err := m.c.IAMRole.Query().Where(iamrole.RoleCodeEQ(roleCode)).Only(ctx)
	if err != nil {
		return err
	}
	_, err = m.c.IAMUser.UpdateOneID(userEnt.ID).AddRoleIDs(roleEnt.ID).Save(ctx)
	return err
}

// RemoveUserRole 解绑角色。
func (m *IAM) RemoveUserRole(ctx context.Context, userCode, roleCode string) error {
	userEnt, err := m.c.IAMUser.Query().Where(iamuser.UserCodeEQ(userCode)).Only(ctx)
	if err != nil {
		return err
	}
	roleEnt, err := m.c.IAMRole.Query().Where(iamrole.RoleCodeEQ(roleCode)).Only(ctx)
	if err != nil {
		return err
	}
	_, err = m.c.IAMUser.UpdateOneID(userEnt.ID).RemoveRoleIDs(roleEnt.ID).Save(ctx)
	return err
}

// IsNotFound 透出 ent NotFound 判断，供上层翻译错误。
func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}
