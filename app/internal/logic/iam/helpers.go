package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
)

const avatarURLExpire = 10 * time.Minute

// currentUserCode 从 JWT 上下文取当前登录用户的 user_code（claim 由登录逻辑写入）。
func currentUserCode(ctx context.Context) string {
	if v := ctx.Value("user_code"); v != nil {
		if s := stringifyClaim(v); s != "" {
			return s
		}
	}
	if v := ctx.Value("sub"); v != nil {
		return stringifyClaim(v)
	}
	return ""
}

func stringifyClaim(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case json.Number:
		return strings.TrimSpace(x.String())
	case fmt.Stringer:
		return strings.TrimSpace(x.String())
	default:
		s := strings.TrimSpace(fmt.Sprint(x))
		if s == "" || s == "<nil>" {
			return ""
		}
		return s
	}
}

// resolveAvatarURL 返回头像可访问 URL：优先按对象重签（私有桶），否则回退快照。
func resolveAvatarURL(ctx context.Context, svcCtx *svc.ServiceContext, objectID uint64, rawURL string) string {
	if objectID == 0 {
		return rawURL
	}
	obj, err := svcCtx.Object.GetByID(ctx, objectID)
	if err != nil || obj == nil {
		return rawURL
	}
	if svcCtx.Cos == nil {
		return obj.URL
	}
	signed, err := svcCtx.Cos.PresignedGetURL(ctx, obj.Key, avatarURLExpire)
	if err != nil || signed == "" {
		return obj.URL
	}
	return signed
}

// avatarSnapshotURL 取对象的规范 URL 作为落库快照（objectID 为 0 返回空串）。
func avatarSnapshotURL(ctx context.Context, svcCtx *svc.ServiceContext, objectID uint64) (string, error) {
	if objectID == 0 {
		return "", nil
	}
	obj, err := svcCtx.Object.GetByID(ctx, objectID)
	if err != nil {
		return "", err
	}
	return obj.URL, nil
}

// toUserItem 将 svc 层用户明细映射为 API 类型（含头像重签）。
func toUserItem(ctx context.Context, svcCtx *svc.ServiceContext, row *svc.IAMUserListItem) types.IAMUserItem {
	roles := row.Roles
	if roles == nil {
		roles = []string{}
	}
	return types.IAMUserItem{
		Id:             row.ID,
		UserCode:       row.UserCode,
		DisplayName:    row.DisplayName,
		Phone:          row.Phone,
		Roles:          roles,
		AvatarObjectId: row.AvatarObjectID,
		AvatarUrl:      resolveAvatarURL(ctx, svcCtx, row.AvatarObjectID, row.AvatarURL),
		HasPassword:    row.HasPassword,
	}
}
