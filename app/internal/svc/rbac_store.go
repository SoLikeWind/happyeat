package svc

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/solikewind/happyeat/app/internal/pkg/casbinrules"
	"github.com/solikewind/happyeat/dal/model/ent"
	iamdal "github.com/solikewind/happyeat/dal/model/iam"
	"golang.org/x/crypto/bcrypt"
)

// RbacPolicyRule 与 Casbin 投影中的 (obj, act) 一致，定义见 casbinrules.PolicyRule。
type RbacPolicyRule = casbinrules.PolicyRule

var roleCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,31}$`)

// phonePattern 中国大陆手机号（简单校验：1 开头 11 位）。
var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// RbacStore 承载 IAM/RBAC 的服务逻辑（校验、Casbin 种子/投影编排），
// 纯数据访问委托给 dal/model/iam。
type RbacStore struct {
	mu  sync.RWMutex
	iam *iamdal.IAM
}

// NewRbacStore 基于 ent client 构造（内部封装 IAM DAL）。
func NewRbacStore(client *ent.Client) (*RbacStore, error) {
	store := &RbacStore{iam: iamdal.NewIAM(client)}
	if err := store.bootstrap(context.Background()); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *RbacStore) List() (map[string][]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.iam.RolePermissionsMap(context.Background())
}

func (s *RbacStore) UpdateRole(roleCode string, permissions []string) error {
	seen := map[string]struct{}{}
	dedup := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		if _, ok := casbinrules.ValidPermissions[permission]; !ok {
			return errors.New("invalid permission: " + permission)
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		dedup = append(dedup, permission)
	}
	sort.Strings(dedup)

	err := s.iam.SetRolePermissions(context.Background(), roleCode, dedup)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("role not found")
		}
		if errors.Is(err, iamdal.ErrPermissionsNotFound) {
			return errors.New("permissions not found")
		}
		return err
	}
	return nil
}

func (s *RbacStore) ListUserRoles() (map[string][]string, error) {
	return s.iam.UserRolesMap(context.Background())
}

func (s *RbacStore) AssignUserRole(userCode, roleCode string) error {
	if err := s.requireRole(roleCode); err != nil {
		return err
	}
	ctx := context.Background()
	exists, err := s.iam.UserExistsByCode(ctx, userCode)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("用户不存在，请先创建用户")
	}
	hasRole, err := s.iam.UserHasRole(ctx, userCode, roleCode)
	if err != nil {
		return err
	}
	if hasRole {
		return nil
	}
	return s.iam.AddUserRole(ctx, userCode, roleCode)
}

// RemoveUserRole 解除用户与角色的关联；用户不存在返回错误，未绑定则幂等成功。
func (s *RbacStore) RemoveUserRole(userCode, roleCode string) error {
	if strings.TrimSpace(userCode) == "" {
		return errors.New("user_code 不能为空")
	}
	if err := s.requireRole(roleCode); err != nil {
		return err
	}
	ctx := context.Background()
	exists, err := s.iam.UserExistsByCode(ctx, userCode)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("用户未找到")
	}
	hasRole, err := s.iam.UserHasRole(ctx, userCode, roleCode)
	if err != nil {
		return err
	}
	if !hasRole {
		return nil
	}
	return s.iam.RemoveUserRole(ctx, userCode, roleCode)
}

// IAMPermissionListItem 分页列出权限点（供 IAM API 使用）。
type IAMPermissionListItem struct {
	Code        string
	Description string
}

// ListIAMPermissionsPage 按 keyword 模糊匹配 permission_code / description，分页升序 code。
func (s *RbacStore) ListIAMPermissionsPage(ctx context.Context, offset, limit int, keyword string) ([]IAMPermissionListItem, int64, error) {
	rows, total, err := s.iam.ListPermissionsPage(ctx, offset, limit, strings.TrimSpace(keyword))
	if err != nil {
		return nil, 0, err
	}
	out := make([]IAMPermissionListItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, IAMPermissionListItem{
			Code:        row.PermissionCode,
			Description: row.Description,
		})
	}
	return out, int64(total), nil
}

// IAMRoleListItem 分页列出角色。
type IAMRoleListItem struct {
	ID       uint64
	RoleCode string
	RoleName string
}

// ListIAMRolesPage 按 keyword 模糊匹配 role_code / role_name。
func (s *RbacStore) ListIAMRolesPage(ctx context.Context, offset, limit int, keyword string) ([]IAMRoleListItem, int64, error) {
	rows, total, err := s.iam.ListRolesPage(ctx, offset, limit, strings.TrimSpace(keyword))
	if err != nil {
		return nil, 0, err
	}
	out := make([]IAMRoleListItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, IAMRoleListItem{
			ID:       row.ID,
			RoleCode: row.RoleCode,
			RoleName: row.RoleName,
		})
	}
	return out, int64(total), nil
}

// CreateRole 创建自定义角色（role_code 创建后不可改）。
func (s *RbacStore) CreateRole(roleCode, roleName string) (uint64, error) {
	roleCode = strings.TrimSpace(strings.ToLower(roleCode))
	if !roleCodePattern.MatchString(roleCode) {
		return 0, errors.New("role_code 须为小写字母、数字、下划线，且以字母开头，长度 2~32")
	}
	if roleName == "" {
		roleName = roleCode
	}
	ctx := context.Background()
	exists, err := s.iam.RoleExists(ctx, roleCode)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("role_code 已存在")
	}
	row, err := s.iam.CreateRole(ctx, roleCode, roleName)
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

// DeleteRoleByID 软删角色；预置角色不可删。
func (s *RbacStore) DeleteRoleByID(id uint64) error {
	ctx := context.Background()
	row, err := s.iam.GetRoleByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("角色不存在")
		}
		return err
	}
	if casbinrules.IsPresetRole(row.RoleCode) {
		return errors.New("系统预置角色不可删除")
	}
	return s.iam.DeleteRoleByID(ctx, id)
}

// IAMUserListItem 分页列出用户及其角色 code。
type IAMUserListItem struct {
	ID             uint64
	UserCode       string
	Phone          string
	DisplayName    string
	Roles          []string
	AvatarObjectID uint64
	AvatarURL      string
	HasPassword    bool
}

// GetUserRoleCodes 返回用户已绑定的角色编码（升序）。
func (s *RbacStore) GetUserRoleCodes(userCode string) ([]string, error) {
	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return nil, errors.New("user_code 不能为空")
	}
	row, err := s.iam.GetByCode(context.Background(), userCode, true)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	out := make([]string, 0, len(row.Edges.Roles))
	for _, r := range row.Edges.Roles {
		out = append(out, r.RoleCode)
	}
	return out, nil
}

// CreateUser 创建 IAM 用户主体（手机号唯一 + bcrypt 密码；角色另行 AssignUserRole）。
// user_code 留空时自动取手机号作为稳定内部主体。
func (s *RbacStore) CreateUser(userCode, displayName, phone, password string) (uint64, error) {
	phone = strings.TrimSpace(phone)
	if !phonePattern.MatchString(phone) {
		return 0, errors.New("手机号格式不正确")
	}
	if len(password) < 6 {
		return 0, errors.New("密码至少 6 位")
	}
	userCode = strings.TrimSpace(strings.ToLower(userCode))
	if userCode == "" {
		userCode = phone
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = phone
	}
	hash, err := hashPassword(password)
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	if exists, err := s.iam.PhoneExists(ctx, phone); err != nil {
		return 0, err
	} else if exists {
		return 0, errors.New("手机号已被注册")
	}
	if exists, err := s.iam.UserExistsByCode(ctx, userCode); err != nil {
		return 0, err
	} else if exists {
		return 0, errors.New("user_code 已存在")
	}

	row, err := s.iam.Create(ctx, iamdal.CreateRow{
		UserCode:     userCode,
		Phone:        phone,
		DisplayName:  displayName,
		PasswordHash: hash,
	})
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

// LoginUser 登录校验主体（供 loginlogic 使用）。
type LoginUser struct {
	UserCode    string
	DisplayName string
	Phone       string
	AvatarURL   string
	Roles       []string
}

// AuthenticateByPhone 用手机号 + 密码校验登录；成功返回主体与角色。
func (s *RbacStore) AuthenticateByPhone(phone, password string) (*LoginUser, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, errors.New("手机号不能为空")
	}
	row, err := s.iam.GetByPhone(context.Background(), phone, true)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	return finishAuth(row, password)
}

// AuthenticateByUserCode 用 user_code + 密码校验登录（兼容 dev-admin 等非手机号账号）。
func (s *RbacStore) AuthenticateByUserCode(userCode, password string) (*LoginUser, error) {
	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return nil, errors.New("账号不能为空")
	}
	row, err := s.iam.GetByCode(context.Background(), userCode, true)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	return finishAuth(row, password)
}

func finishAuth(row *ent.IAMUser, password string) (*LoginUser, error) {
	if row.PasswordHash == "" {
		return nil, errors.New("该账号未设置密码，请联系管理员")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	roles := make([]string, 0, len(row.Edges.Roles))
	for _, r := range row.Edges.Roles {
		roles = append(roles, r.RoleCode)
	}
	if len(roles) == 0 {
		return nil, errors.New("该用户未分配角色，请联系管理员")
	}
	return &LoginUser{
		UserCode:    row.UserCode,
		DisplayName: row.DisplayName,
		Phone:       row.Phone,
		AvatarURL:   row.AvatarURL,
		Roles:       roles,
	}, nil
}

// SetPassword 设置/重置用户密码（bcrypt）。
func (s *RbacStore) SetPassword(userCode, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("密码至少 6 位")
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	err = s.iam.UpdatePasswordByCode(context.Background(), userCode, hash)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("用户不存在")
		}
		return err
	}
	return nil
}

// SetPasswordByID 按用户 ID 重置密码（管理员用）。
func (s *RbacStore) SetPasswordByID(id uint64, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("密码至少 6 位")
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	ctx := context.Background()
	exists, err := s.iam.UserExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("用户不存在")
	}
	return s.iam.UpdatePasswordByID(ctx, id, hash)
}

// ChangePassword 用户自助改密：校验旧密码后设置新密码。
func (s *RbacStore) ChangePassword(userCode, oldPassword, newPassword string) error {
	if _, err := s.AuthenticateByUserCode(userCode, oldPassword); err != nil {
		return errors.New("原密码不正确")
	}
	return s.SetPassword(userCode, newPassword)
}

// SetAvatar 更新用户头像对象与 URL 快照。
func (s *RbacStore) SetAvatar(userCode string, objectID uint64, url string) error {
	err := s.iam.UpdateAvatarByCode(context.Background(), userCode, objectID, url)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("用户不存在")
		}
		return err
	}
	return nil
}

// UpdateUserByID 更新用户展示名/手机号/头像（管理员用，nil 表示不改）。
func (s *RbacStore) UpdateUserByID(id uint64, displayName, phone *string, avatarObjectID *uint64, avatarURL *string) error {
	ctx := context.Background()
	exists, err := s.iam.UserExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("用户不存在")
	}
	upd := iamdal.UpdateRow{AvatarObjectID: avatarObjectID, AvatarURL: avatarURL}
	if displayName != nil {
		dn := strings.TrimSpace(*displayName)
		upd.DisplayName = &dn
	}
	if phone != nil {
		p := strings.TrimSpace(*phone)
		if !phonePattern.MatchString(p) {
			return errors.New("手机号格式不正确")
		}
		dup, err := s.iam.PhoneExistsExceptID(ctx, p, id)
		if err != nil {
			return err
		}
		if dup {
			return errors.New("手机号已被注册")
		}
		upd.Phone = &p
	}
	return s.iam.UpdateByID(ctx, id, upd)
}

// GetUserDetailByCode 读取用户明细（含角色、头像对象），供 me/详情接口。
func (s *RbacStore) GetUserDetailByCode(userCode string) (*IAMUserListItem, error) {
	row, err := s.iam.GetByCode(context.Background(), userCode, true)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return entUserToListItem(row), nil
}

// GetUserDetailByID 读取用户明细（管理员详情接口）。
func (s *RbacStore) GetUserDetailByID(id uint64) (*IAMUserListItem, error) {
	row, err := s.iam.GetByID(context.Background(), id, true)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return entUserToListItem(row), nil
}

func entUserToListItem(row *ent.IAMUser) *IAMUserListItem {
	roleCodes := make([]string, 0, len(row.Edges.Roles))
	for _, r := range row.Edges.Roles {
		roleCodes = append(roleCodes, r.RoleCode)
	}
	return &IAMUserListItem{
		ID:             row.ID,
		UserCode:       row.UserCode,
		Phone:          row.Phone,
		DisplayName:    row.DisplayName,
		Roles:          roleCodes,
		AvatarObjectID: row.AvatarObjectID,
		AvatarURL:      row.AvatarURL,
		HasPassword:    row.PasswordHash != "",
	}
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DeleteUserByID 软删用户并清除角色关联。
func (s *RbacStore) DeleteUserByID(id uint64) error {
	ctx := context.Background()
	exists, err := s.iam.UserExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("用户不存在")
	}
	return s.iam.DeleteByID(ctx, id)
}

// ListIAMUsersPage 按 keyword 模糊匹配 user_code / display_name。
func (s *RbacStore) ListIAMUsersPage(ctx context.Context, offset, limit int, keyword string) ([]IAMUserListItem, int64, error) {
	rows, total, err := s.iam.ListUsersPage(ctx, offset, limit, strings.TrimSpace(keyword))
	if err != nil {
		return nil, 0, err
	}
	out := make([]IAMUserListItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, *entUserToListItem(row))
	}
	return out, int64(total), nil
}

func (s *RbacStore) Reset(roleCode string) error {
	defaults := defaultRolePermissions()
	if roleCode == "" {
		for rc, permissions := range defaults {
			if err := s.UpdateRole(rc, permissions); err != nil {
				return err
			}
		}
		return nil
	}
	permissions, ok := defaults[roleCode]
	if !ok {
		return s.UpdateRole(roleCode, []string{})
	}
	return s.UpdateRole(roleCode, permissions)
}

func (s *RbacStore) bootstrap(ctx context.Context) error {
	if err := s.seedRolesAndPermissions(ctx); err != nil {
		return err
	}
	if err := s.seedDefaultMappings(); err != nil {
		return err
	}
	return nil
}

func (s *RbacStore) seedRolesAndPermissions(ctx context.Context) error {
	defaults := defaultRolePermissions()
	presetNames := map[string]string{
		"super_admin": "超级管理员",
		"manager":     "店长",
		"cashier":     "收银",
		"kitchen":     "后厨",
		"waiter":      "服务员",
	}
	for roleCode := range defaults {
		name := presetNames[roleCode]
		if name == "" {
			name = roleCode
		}
		if err := s.iam.EnsureRole(ctx, roleCode, name); err != nil {
			return err
		}
	}
	for _, permission := range casbinrules.PermissionCatalog {
		if err := s.iam.EnsurePermission(ctx, permission.Code, permission.Description); err != nil {
			return err
		}
	}
	return nil
}

func (s *RbacStore) seedDefaultMappings() error {
	ctx := context.Background()
	count, err := s.iam.RolePermissionsCount(ctx)
	if err != nil {
		return err
	}
	if count == 0 {
		defaults := defaultRolePermissions()
		for roleCode, permissions := range defaults {
			if err := s.UpdateRole(roleCode, permissions); err != nil {
				return err
			}
		}
	}
	// dev-admin：开发态内置超管账号（占位手机号 + 默认密码 admin123），用 user_code 登录。
	if err := s.ensureDevAdmin(ctx); err != nil {
		return err
	}
	hasRole, err := s.iam.UserHasRole(ctx, "dev-admin", "super_admin")
	if err != nil {
		return err
	}
	if !hasRole {
		if err := s.AssignUserRole("dev-admin", "super_admin"); err != nil {
			return err
		}
	}
	return nil
}

// ensureDevAdmin 确保开发态超管账号存在且已设密码。
func (s *RbacStore) ensureDevAdmin(ctx context.Context) error {
	const code = "dev-admin"
	row, err := s.iam.GetByCode(ctx, code, false)
	if err != nil {
		if !iamdal.IsNotFound(err) {
			return err
		}
		hash, herr := hashPassword("admin123")
		if herr != nil {
			return herr
		}
		_, cerr := s.iam.Create(ctx, iamdal.CreateRow{
			UserCode:     code,
			Phone:        "admin",
			DisplayName:  "超级管理员",
			PasswordHash: hash,
		})
		return cerr
	}
	if row.PasswordHash == "" {
		hash, herr := hashPassword("admin123")
		if herr != nil {
			return herr
		}
		return s.iam.UpdatePasswordByID(ctx, row.ID, hash)
	}
	return nil
}

func (s *RbacStore) requireRole(roleCode string) error {
	exists, err := s.iam.RoleExists(context.Background(), roleCode)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("role not found")
	}
	return nil
}

func BuildPoliciesForPermissions(permissions []string) []RbacPolicyRule {
	out := make([]RbacPolicyRule, 0, len(permissions)*2)
	uniq := make(map[string]struct{})
	for _, permission := range permissions {
		rules := casbinrules.PermissionRules[permission]
		for _, rule := range rules {
			key := rule.Act + "|" + rule.Obj
			if _, ok := uniq[key]; ok {
				continue
			}
			uniq[key] = struct{}{}
			out = append(out, rule)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Obj == out[j].Obj {
			return out[i].Act < out[j].Act
		}
		return out[i].Obj < out[j].Obj
	})
	return out
}

func defaultRolePermissions() map[string][]string {
	all := make([]string, 0, len(casbinrules.ValidPermissions))
	for permission := range casbinrules.ValidPermissions {
		all = append(all, permission)
	}
	sort.Strings(all)
	return map[string][]string{
		"super_admin": append([]string{}, all...),
		"manager":     append([]string{}, all...),
		"cashier":     {"home:view", "orders:view", "orders:create", "orders:print_kitchen", "order_desk:view", "order_desk:create", "stats:view", "spec:view", "settlements:view", "settlements:edit", "settlements:settle"},
		"kitchen":     {"home:view", "workbench:view", "workbench:complete", "orders:view", "orders:print_kitchen", "spec:view"},
		"waiter":      {"home:view", "orders:view", "orders:print_kitchen", "order_desk:view", "order_desk:create", "table:view", "spec:view"},
		"unknown":     {},
	}
}
