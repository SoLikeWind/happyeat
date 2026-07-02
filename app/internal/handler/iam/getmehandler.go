// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"net/http"

	"github.com/solikewind/happyeat/app/internal/logic/iam"
	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取当前登录用户资料（含头像、角色）
func GetMeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := iam.NewGetMeLogic(r.Context(), svcCtx)
		resp, err := l.GetMe()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
