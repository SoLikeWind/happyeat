// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeMyPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 当前登录用户修改密码
func NewChangeMyPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeMyPasswordLogic {
	return &ChangeMyPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangeMyPasswordLogic) ChangeMyPassword(req *types.ChangeMyPasswordReq) (resp *types.ChangeMyPasswordReply, err error) {
	userCode := currentUserCode(l.ctx)
	if userCode == "" {
		return nil, errInvalid("登录状态无效，请重新登录")
	}
	if err := l.svcCtx.Rbac.ChangePassword(userCode, req.OldPassword, req.NewPassword); err != nil {
		return nil, errInvalid(err.Error())
	}
	return &types.ChangeMyPasswordReply{}, nil
}
