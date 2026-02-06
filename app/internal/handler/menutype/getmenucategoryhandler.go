// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menutype

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"happyeat/app/internal/logic/menutype"
	"happyeat/app/internal/svc"
	"happyeat/app/internal/types"
)

// 获取菜单种类
func GetMenuCategoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMenuCategoryReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := menutype.NewGetMenuCategoryLogic(r.Context(), svcCtx)
		resp, err := l.GetMenuCategory(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
