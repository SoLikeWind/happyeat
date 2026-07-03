// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetIAMPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取单个权限点
func NewGetIAMPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetIAMPermissionLogic {
	return &GetIAMPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetIAMPermissionLogic) GetIAMPermission(req *types.GetIAMPermissionReq) (resp *types.GetIAMPermissionReply, err error) {
	if req.Id == 0 {
		return nil, errInvalid("id 不能为空")
	}
	permission, err := l.svcCtx.Rbac.GetIAMPermissionByID(req.Id)
	if err != nil {
		return nil, errInvalid(err.Error())
	}
	return &types.GetIAMPermissionReply{
		Permission: types.PermissionItem{
			Id:          permission.ID,
			Code:        permission.Code,
			Description: permission.Description,
		},
	}, nil
}
