// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前登录用户资料（含头像、角色）
func NewGetMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMeLogic {
	return &GetMeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMeLogic) GetMe() (resp *types.GetMeReply, err error) {
	userCode := currentUserCode(l.ctx)
	if userCode == "" {
		return nil, errInvalid("登录状态无效，请重新登录")
	}
	detail, err := l.svcCtx.Rbac.GetUserDetailByCode(userCode)
	if err != nil {
		return nil, errInvalid(err.Error())
	}
	return &types.GetMeReply{User: toUserItem(l.ctx, l.svcCtx, detail)}, nil
}
