// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package audit

import (
	"context"

	"github.com/solikewind/happyeat/app/internal/svc"
	"github.com/solikewind/happyeat/app/internal/types"
	auditmodel "github.com/solikewind/happyeat/dal/model/audit"
	"github.com/solikewind/happyeat/dal/model/ent"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOperationLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 分页查询后台操作日志
func NewListOperationLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOperationLogsLogic {
	return &ListOperationLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOperationLogsLogic) ListOperationLogs(req *types.ListOperationLogsReq) (resp *types.ListOperationLogsReply, err error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	current := int(req.Current)
	if current <= 0 {
		current = 1
	}
	list, total, err := l.svcCtx.Audit.ListOperationLogsPage(l.ctx, auditmodel.ListOperationLogsFilter{
		ActorUserCode: req.ActorUserCode,
		Module:        req.Module,
		Action:        req.Action,
		Method:        req.Method,
		Keyword:       req.Keyword,
		Offset:        (current - 1) * pageSize,
		Limit:         pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.OperationLogItem, 0, len(list))
	for _, row := range list {
		items = append(items, toOperationLogItem(row))
	}
	return &types.ListOperationLogsReply{
		Logs:  items,
		Total: total,
	}, nil
}

func toOperationLogItem(row *ent.OperationLog) types.OperationLogItem {
	if row == nil {
		return types.OperationLogItem{}
	}
	return types.OperationLogItem{
		Id:             row.ID,
		ActorUserCode:  row.ActorUserCode,
		Module:         row.Module,
		Action:         row.Action,
		Method:         row.Method,
		Path:           row.Path,
		NormalizedPath: row.NormalizedPath,
		Status:         row.Status,
		TargetId:       row.TargetID,
		Ip:             row.IP,
		UserAgent:      row.UserAgent,
		Summary:        row.Summary,
		CreatedAt:      row.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
