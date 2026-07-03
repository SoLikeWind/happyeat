// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetIAMRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取单个角色
func NewGetIAMRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetIAMRoleLogic {
	return &GetIAMRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetIAMRoleLogic) GetIAMRole(req *types.GetIAMRoleReq) (resp *types.GetIAMRoleReply, err error) {
	if req.Id == 0 {
		return nil, errInvalid("id 不能为空")
	}
	role, err := l.svcCtx.Rbac.GetIAMRoleByID(req.Id)
	if err != nil {
		return nil, errInvalid(err.Error())
	}
	return &types.GetIAMRoleReply{
		Role: types.IAMRoleItem{
			Id:       role.ID,
			RoleCode: role.RoleCode,
			RoleName: role.RoleName,
		},
	}, nil
}
