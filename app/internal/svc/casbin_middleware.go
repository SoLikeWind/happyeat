package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/solikewind/happyeat/app/internal/pkg/routenorm"
	auditmodel "github.com/solikewind/happyeat/dal/model/audit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func NewCasbinMiddleware(svcCtx *ServiceContext) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path) {
				next(w, r)
				return
			}

			sub := extractSubject(r)
			if sub == "" {
				writeUnauthorizedIdentity(w)
				return
			}

			// 自助接口（/iam/me 等）：已登录即放行，不做 Casbin 策略校验。
			if routenorm.IsSelfServicePath(r.URL.Path) {
				serveWithAudit(svcCtx, sub, next, w, r)
				return
			}

			obj := routenorm.EnforceObj(r.URL.Path)
			act := strings.ToUpper(r.Method)
			ok, err := svcCtx.Casbin.Enforcer.Enforce(sub, obj, act)
			if err != nil {
				writeCasbinInternalError(w)
				return
			}
			if !ok {
				writeForbidden(w, "forbidden: insufficient permissions")
				return
			}

			serveWithAudit(svcCtx, sub, next, w, r)
		}
	}
}

func serveWithAudit(svcCtx *ServiceContext, actor string, next http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	next(rec, r)
	writeAuditLog(svcCtx, actor, r, rec.status)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func writeAuditLog(svcCtx *ServiceContext, actor string, r *http.Request, status int) {
	if svcCtx == nil || svcCtx.Audit == nil || !shouldAuditMethod(r.Method) {
		return
	}
	normalized := routenorm.EnforceObj(r.URL.Path)
	module := auditModule(normalized)
	action := auditAction(r.Method)
	targetID := auditTargetID(r.URL.Path)
	summary := fmt.Sprintf("%s %s %s", action, module, normalized)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := svcCtx.Audit.CreateOperationLog(ctx, auditmodel.CreateOperationLogInput{
		ActorUserCode:  actor,
		Module:         module,
		Action:         action,
		Method:         r.Method,
		Path:           r.URL.Path,
		NormalizedPath: normalized,
		Status:         status,
		TargetID:       targetID,
		IP:             clientIP(r),
		UserAgent:      r.UserAgent(),
		Summary:        summary,
	}); err != nil {
		logx.Errorf("write operation log failed: %v", err)
	}
}

func shouldAuditMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return true
	default:
		return false
	}
}

func auditAction(method string) string {
	switch strings.ToUpper(method) {
	case http.MethodPost:
		return "创建/执行"
	case http.MethodPut, http.MethodPatch:
		return "更新"
	case http.MethodDelete:
		return "删除"
	default:
		return strings.ToUpper(method)
	}
}

func auditModule(normalizedPath string) string {
	parts := strings.Split(strings.Trim(normalizedPath, "/"), "/")
	if len(parts) >= 3 && parts[0] == "central" && parts[1] == "v1" {
		return moduleLabel(parts[2])
	}
	return "系统"
}

func moduleLabel(segment string) string {
	switch segment {
	case "iam", "rbac", "operation-logs":
		return "权限与账号"
	case "menus", "menu", "objects", "object", "spec":
		return "菜单"
	case "tables", "table":
		return "餐桌"
	case "orders", "order", "workbench":
		return "订单"
	case "settlements", "settlement":
		return "结账单"
	case "stats":
		return "统计"
	default:
		return segment
	}
}

func auditTargetID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.TrimSpace(parts[i])
		if p == "" {
			continue
		}
		if _, err := strconv.ParseUint(p, 10, 64); err == nil {
			return p
		}
	}
	return ""
}

func clientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); v != "" {
		return strings.TrimSpace(strings.Split(v, ",")[0])
	}
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		return v
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func isPublicPath(path string) bool {
	switch path {
	case "/health", "/openapi/happyeat.json", "/central/v1/auth/login":
		return true
	default:
		return false
	}
}

func extractSubject(r *http.Request) string {
	if s := stringFromContext(r, "user_code"); s != "" {
		return s
	}
	if userClaims, ok := r.Context().Value("user").(map[string]any); ok {
		if sub, ok := userClaims["sub"].(string); ok && strings.TrimSpace(sub) != "" {
			return strings.TrimSpace(sub)
		}
		if role, ok := userClaims["role"].(string); ok && strings.TrimSpace(role) != "" {
			return strings.TrimSpace(role)
		}
	}
	return ""
}

func stringFromContext(r *http.Request, key string) string {
	v := r.Context().Value(key)
	if v == nil {
		return ""
	}
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

func writeUnauthorizedIdentity(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": 401,
		"msg":  "登录已失效或令牌格式过旧，请重新登录",
	})
}

func writeCasbinInternalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": 500,
		"msg":  "权限服务暂时不可用",
	})
}

func writeForbidden(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": 403,
		"msg":  msg,
	})
}
