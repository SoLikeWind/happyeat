// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package audit

import (
	"net/http"

	"github.com/solikewind/happyeat/app/internal/logic/audit"
	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 分页查询后台操作日志
func ListOperationLogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListOperationLogsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := audit.NewListOperationLogsLogic(r.Context(), svcCtx)
		resp, err := l.ListOperationLogs(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
