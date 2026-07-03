package audit

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/solikewind/happyeat/dal/model/ent"
	"github.com/solikewind/happyeat/dal/model/ent/operationlog"
)

type Audit struct {
	c *ent.Client
}

func NewAudit(c *ent.Client) *Audit {
	return &Audit{c: c}
}

type CreateOperationLogInput struct {
	ActorUserCode  string
	Module         string
	Action         string
	Method         string
	Path           string
	NormalizedPath string
	Status         int
	TargetID       string
	IP             string
	UserAgent      string
	Summary        string
}

func (a *Audit) CreateOperationLog(ctx context.Context, in CreateOperationLogInput) error {
	if strings.TrimSpace(in.ActorUserCode) == "" {
		return nil
	}
	return a.c.OperationLog.Create().
		SetActorUserCode(trimMax(in.ActorUserCode, 128)).
		SetModule(trimMax(in.Module, 64)).
		SetAction(trimMax(in.Action, 64)).
		SetMethod(trimMax(strings.ToUpper(in.Method), 16)).
		SetPath(trimMax(in.Path, 512)).
		SetNormalizedPath(trimMax(in.NormalizedPath, 512)).
		SetStatus(in.Status).
		SetTargetID(trimMax(in.TargetID, 128)).
		SetIP(trimMax(in.IP, 128)).
		SetUserAgent(trimMax(in.UserAgent, 512)).
		SetSummary(trimMax(in.Summary, 512)).
		Exec(ctx)
}

type ListOperationLogsFilter struct {
	ActorUserCode string
	Module        string
	Action        string
	Method        string
	Keyword       string
	Offset        int
	Limit         int
}

func (a *Audit) ListOperationLogsPage(ctx context.Context, f ListOperationLogsFilter) ([]*ent.OperationLog, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 10
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	q := a.c.OperationLog.Query()
	if s := strings.TrimSpace(f.ActorUserCode); s != "" {
		q = q.Where(operationlog.ActorUserCodeContainsFold(s))
	}
	if s := strings.TrimSpace(f.Module); s != "" {
		q = q.Where(operationlog.ModuleEQ(s))
	}
	if s := strings.TrimSpace(f.Action); s != "" {
		q = q.Where(operationlog.ActionEQ(s))
	}
	if s := strings.TrimSpace(f.Method); s != "" {
		q = q.Where(operationlog.MethodEQ(strings.ToUpper(s)))
	}
	if s := strings.TrimSpace(f.Keyword); s != "" {
		q = q.Where(operationlog.Or(
			operationlog.PathContainsFold(s),
			operationlog.NormalizedPathContainsFold(s),
			operationlog.TargetIDContainsFold(s),
			operationlog.SummaryContainsFold(s),
		))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	list, err := q.
		Order(operationlog.ByCreatedAt(sql.OrderDesc()), operationlog.ByID(sql.OrderDesc())).
		Offset(f.Offset).
		Limit(f.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, int64(total), nil
}

func trimMax(s string, max int) string {
	s = strings.TrimSpace(s)
	rs := []rune(s)
	if len(rs) <= max {
		return s
	}
	return string(rs[:max])
}
