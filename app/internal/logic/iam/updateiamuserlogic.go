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

type UpdateIAMUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新用户展示名
func NewUpdateIAMUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIAMUserLogic {
	return &UpdateIAMUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateIAMUserLogic) UpdateIAMUser(req *types.UpdateIAMUserReq) (resp *types.UpdateIAMUserReply, err error) {
	if req.Id == 0 {
		return nil, errInvalid("id 不能为空")
	}

	var displayName, phone, avatarURL *string
	var avatarObjectID *uint64

	if s := strings.TrimSpace(req.DisplayName); s != "" {
		displayName = &s
	}
	if s := strings.TrimSpace(req.Phone); s != "" {
		phone = &s
	}
	// 头像按“全量意图”处理：0 表示清除，非 0 表示设置为该对象（前端编辑弹窗始终回填当前头像，不会误清）。
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

	if err := l.svcCtx.Rbac.UpdateUserByID(req.Id, displayName, phone, avatarObjectID, avatarURL); err != nil {
		return nil, errInvalid(err.Error())
	}
	return &types.UpdateIAMUserReply{}, nil
}
