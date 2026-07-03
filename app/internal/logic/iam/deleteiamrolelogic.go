// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteIAMRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除角色（软删；系统预置角色不可删；删除后全量同步 Casbin）
func NewDeleteIAMRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteIAMRoleLogic {
	return &DeleteIAMRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteIAMRoleLogic) DeleteIAMRole(req *types.DeleteIAMRoleReq) (resp *types.DeleteIAMRoleReply, err error) {
	if req.Id == 0 {
		return nil, errInvalid("id 不能为空")
	}
	outcome, err := l.svcCtx.Rbac.DeleteRoleByID(req.Id)
	if err != nil {
		return nil, errInvalid(err.Error())
	}
	if err := svc.RemoveRolePoliciesFromCasbin(l.svcCtx.Casbin, outcome.RoleCode, outcome.UserCodes, outcome.Permissions); err != nil {
		return nil, errInvalid("同步 Casbin 策略失败")
	}
	return &types.DeleteIAMRoleReply{}, nil
}
