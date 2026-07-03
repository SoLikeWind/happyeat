// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package iam

import (
	"context"
	"strings"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateIAMUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建用户（仅主体档案，分配角色用 user-roles 接口）
func NewCreateIAMUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateIAMUserLogic {
	return &CreateIAMUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateIAMUserLogic) CreateIAMUser(req *types.CreateIAMUserReq) (resp *types.CreateIAMUserReply, err error) {
	userCode := strings.TrimSpace(strings.ToLower(req.UserCode))
	displayName := strings.TrimSpace(req.DisplayName)
	phone := strings.TrimSpace(req.Phone)
	id, err := l.svcCtx.Rbac.CreateUser(userCode, displayName, phone, req.Password)
	if err != nil {
		return nil, errInvalid(err.Error())
	}

	// user_code 留空时后端以手机号回填，取回真实 user_code 以便绑定角色/设置头像。
	detail, err := l.svcCtx.Rbac.GetUserDetailByID(id)
	if err != nil {
		return nil, errInvalid(err.Error())
	}
	realUserCode := detail.UserCode

	// 可选：一次性绑定角色并同步 Casbin。
	for _, roleCode := range req.Roles {
		rc := strings.TrimSpace(roleCode)
		if rc == "" {
			continue
		}
		if err := l.svcCtx.Rbac.AssignUserRole(realUserCode, rc); err != nil {
			return nil, errInvalid(err.Error())
		}
	}
	if err := svc.EnsureUserCasbinGroupings(l.svcCtx.Rbac, l.svcCtx.Casbin, realUserCode); err != nil {
		return nil, errInvalid("同步 Casbin 用户角色失败")
	}

	// 可选：设置头像。
	if req.AvatarObjectId != 0 {
		url, aerr := avatarSnapshotURL(l.ctx, l.svcCtx, req.AvatarObjectId)
		if aerr != nil {
			return nil, errInvalid("头像对象不存在")
		}
		if err := l.svcCtx.Rbac.SetAvatar(realUserCode, req.AvatarObjectId, url); err != nil {
			return nil, errInvalid(err.Error())
		}
	}

	return &types.CreateIAMUserReply{Id: id}, nil
}
