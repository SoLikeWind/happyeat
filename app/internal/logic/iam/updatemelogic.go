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

type UpdateMeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新当前登录用户资料（展示名/头像）
func NewUpdateMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMeLogic {
	return &UpdateMeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMeLogic) UpdateMe(req *types.UpdateMeReq) (resp *types.UpdateMeReply, err error) {
	userCode := currentUserCode(l.ctx)
	if userCode == "" {
		return nil, errInvalid("登录状态无效，请重新登录")
	}
	detail, err := l.svcCtx.Rbac.GetUserDetailByCode(userCode)
	if err != nil {
		return nil, errInvalid(err.Error())
	}

	var displayName, avatarURL *string
	var avatarObjectID *uint64
	if s := strings.TrimSpace(req.DisplayName); s != "" {
		displayName = &s
	}
	// 头像按“全量意图”处理：0 表示清除，非 0 表示设置为该对象（前端弹窗始终回填当前头像，不会误清）。
	{
		objID := req.AvatarObjectId
		url := ""
		if objID != 0 {
			u, aerr := avatarSnapshotURL(l.ctx, l.svcCtx, objID)
			if aerr != nil {
				return nil, errInvalid("头像对象不存在")
			}
			url = u
		}
		avatarObjectID = &objID
		avatarURL = &url
	}

	// 复用按 ID 更新（禁止本人改手机号，管理员在用户管理页维护）。
	if err := l.svcCtx.Rbac.UpdateUserByID(detail.ID, displayName, nil, avatarObjectID, avatarURL); err != nil {
		return nil, errInvalid(err.Error())
	}
	return &types.UpdateMeReply{}, nil
}
