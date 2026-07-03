// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteIAMUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除用户（软删，并清除 Casbin 中该用户全部分组）
func NewDeleteIAMUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteIAMUserLogic {
	return &DeleteIAMUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteIAMUserLogic) DeleteIAMUser(req *types.DeleteIAMUserReq) (resp *types.DeleteIAMUserReply, err error) {
	if req.Id == 0 {
		return nil, errInvalid("id 不能为空")
	}
	detail, err := l.svcCtx.Rbac.GetUserDetailByID(req.Id)
	if err != nil {
		return nil, errInvalid(err.Error())
	}
	if detail.UserCode == "dev-admin" {
		return nil, errInvalid("内置超级管理员账号不可删除")
	}
	if detail.UserCode == currentUserCode(l.ctx) {
		return nil, errInvalid("不能删除当前登录账号")
	}
	if err := l.svcCtx.Rbac.DeleteUserByID(req.Id); err != nil {
		return nil, errInvalid(err.Error())
	}
	if err := svc.RemoveAllUserCasbinGroupings(l.svcCtx.Casbin, detail.UserCode); err != nil {
		return nil, errInvalid("同步 Casbin 策略失败")
	}
	return &types.DeleteIAMUserReply{}, nil
}
