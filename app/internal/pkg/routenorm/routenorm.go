// Package routenorm 与 Casbin 中间件使用相同的路径归一化规则，供 routecheck 等工具复用。
package routenorm

import (
	"regexp"
	"strings"
)

var idPathSegment = regexp.MustCompile(`/[0-9a-fA-F-]{6,}`)

// numericPathSegmentMiddle / numericPathSegmentEnd 匹配「整段为十进制数字」的路径段（Go regexp 无前瞻，分两段替换）。
var (
	numericPathSegmentMiddle = regexp.MustCompile(`/(\d+)/`)
	numericPathSegmentEnd    = regexp.MustCompile(`/(\d+)$`)
)

const rbacRolePermissionsPrefix = "/central/v1/rbac/role-permissions"

// selfServicePaths 自助接口：已登录用户管理自己资料/头像/改密，无需 Casbin 授权。
// Casbin 中间件放行、routecheck 跳过，二者共用此集合以保持一致。
var selfServicePaths = map[string]struct{}{
	"/central/v1/iam/me":          {},
	"/central/v1/iam/me/password": {},
}

// IsSelfServicePath 判断是否为自助接口（仅需已登录，不做 Casbin 策略校验）。
func IsSelfServicePath(path string) bool {
	_, ok := selfServicePaths[strings.TrimRight(path, "/")]
	return ok
}

// NormalizePath 将请求路径转为 Casbin obj：UUID 段替换为 /:id，并去掉尾部斜杠。
func NormalizePath(path string) string {
	normalized := idPathSegment.ReplaceAllString(path, "/:id")
	return strings.TrimRight(normalized, "/")
}

// EnforceObj 将实际请求 URL 规范为与 casbinrules.PermissionRules 中 p.obj 一致的形式，
// 供 Casbin matcher（r.obj == p.obj）使用：UUID 段、纯数字主键、RBAC 角色路径段均归为 /:id；
// 固定路径如 .../role-permissions/reset 保持不变。
func EnforceObj(path string) string {
	p := NormalizePath(path)
	p = normalizeRbacRolePermissionsPath(p)
	for {
		next := numericPathSegmentMiddle.ReplaceAllString(p, "/:id/")
		next = numericPathSegmentEnd.ReplaceAllString(next, "/:id")
		if next == p {
			break
		}
		p = next
	}
	return p
}

func normalizeRbacRolePermissionsPath(p string) string {
	if !strings.HasPrefix(p, rbacRolePermissionsPrefix+"/") {
		return p
	}
	rest := strings.TrimPrefix(p, rbacRolePermissionsPrefix+"/")
	if rest == "" || strings.Contains(rest, "/") {
		return p
	}
	if rest == "reset" {
		return p
	}
	return rbacRolePermissionsPrefix + "/:id"
}

// CanonicalObj 在 NormalizePath 基础上，将任意命名路径参数（如 :role、:id）统一为 :id，便于路由模板与 permission 规则对比。
func CanonicalObj(path string) string {
	p := NormalizePath(path)
	param := regexp.MustCompile(`:[a-zA-Z_][a-zA-Z0-9_]*`)
	return param.ReplaceAllString(p, ":id")
}
