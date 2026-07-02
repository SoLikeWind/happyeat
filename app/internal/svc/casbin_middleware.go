package svc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/solikewind/happyeat/app/internal/pkg/routenorm"
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
				next(w, r)
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

			next(w, r)
		}
	}
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
