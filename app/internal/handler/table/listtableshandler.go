// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package table

import (
	"net/http"

	"github.com/solikewind/happyeat/app/internal/logic/table"
	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 列出餐桌
func ListTablesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListTablesReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := table.NewListTablesLogic(r.Context(), svcCtx)
		resp, err := l.ListTables(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if resp == nil {
			resp = &types.ListTablesReply{Tables: []types.Table{}, Total: 0}
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
