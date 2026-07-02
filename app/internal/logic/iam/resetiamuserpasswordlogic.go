// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetIAMUserPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员重置用户密码
func NewResetIAMUserPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetIAMUserPasswordLogic {
	return &ResetIAMUserPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetIAMUserPasswordLogic) ResetIAMUserPassword(req *types.ResetIAMUserPasswordReq) (resp *types.ResetIAMUserPasswordReply, err error) {
	if req.Id == 0 {
		return nil, errInvalid("id 不能为空")
	}
	if err := l.svcCtx.Rbac.SetPasswordByID(req.Id, req.Password); err != nil {
		return nil, errInvalid(err.Error())
	}
	return &types.ResetIAMUserPasswordReply{}, nil
}
