// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-kratos/blades"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/solikewind/happyeat/app/internal/config"
	"github.com/solikewind/happyeat/app/internal/pkg/agent"
	"github.com/solikewind/happyeat/app/internal/pkg/cos"
	"github.com/solikewind/happyeat/app/internal/pkg/spyun"
	auditmodel "github.com/solikewind/happyeat/dal/model/audit"
	"github.com/solikewind/happyeat/dal/model/ent"
	_ "github.com/solikewind/happyeat/dal/model/ent/runtime"
	"github.com/solikewind/happyeat/dal/model/menu"
	objmodel "github.com/solikewind/happyeat/dal/model/object"
	"github.com/solikewind/happyeat/dal/model/order"
	"github.com/solikewind/happyeat/dal/model/settlement"
	specmodel "github.com/solikewind/happyeat/dal/model/spec"
	"github.com/solikewind/happyeat/dal/model/table"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config           config.Config
	DB               *sql.DB         // 共享连接池，仅用于关闭
	Casbin           *CasbinEnforcer // 权限：model 来自配置内联，policy 来自 DB casbin_rule 表
	Rbac             *RbacStore
	CasbinMiddleware rest.Middleware
	Agent            *blades.Agent // 智能体
	LLM              *agent.LangChainService
	ASR              *agent.BailianASRClient
	Cos              *cos.Client
	Spyun            *spyun.Client

	Menu         *menu.Menu              // 菜单 data 层
	MenuType     *menu.MenuType          // 菜单分类 data 层
	CategorySpec *specmodel.CategorySpec // 分类规格模板 data 层
	SpecGroup    *specmodel.SpecGroup    // 规格组 data 层
	SpecItem     *specmodel.SpecItem     // 规格项 data 层

	Table     *table.Table     // 餐桌 data 层
	TableType *table.TableType // 餐桌分类 data 层

	Order      *order.Order // 订单 data 层
	Object     *objmodel.Object
	Audit      *auditmodel.Audit
	Settlement *settlement.Settlement // 结账单 data 层
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	db, err := sql.Open("pgx", c.SqlConfig.DataSource)
	if err != nil {
		return nil, err
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	ce, err := NewCasbinEnforcer(c.Casbin.Model, c.SqlConfig.DataSource)
	if err != nil {
		return nil, err
	}

	// 初始化 Agent
	llmConfig, err := agent.NewConfig(c.LLM)
	if err != nil {
		return nil, err
	}
	bladesAgent, err := agent.NewMenusTechAgent(llmConfig, menu.NewMenu(client))
	if err != nil {
		return nil, err
	}

	var llmSvc *agent.LangChainService
	if svc, err := agent.NewLangChainService(c.LLM); err == nil {
		llmSvc = svc
	}

	var asrSvc *agent.BailianASRClient
	if svc, err := agent.NewBailianASRClient(c.ASR); err == nil {
		asrSvc = svc
	}
	rbacStore, err := NewRbacStore(client)
	if err != nil {
		return nil, err
	}

	ctx := &ServiceContext{
		Config: c,
		DB:     db,
		Casbin: ce,
		Rbac:   rbacStore,
		Agent:  bladesAgent.Agent,
		LLM:    llmSvc,
		ASR:    asrSvc,
		Cos:    cos.NewClient(c.Cos),
		Spyun:  spyun.NewClient(c.Spyun),

		Menu:         menu.NewMenu(client),
		MenuType:     menu.NewMenuType(client),
		CategorySpec: specmodel.NewCategorySpec(client),
		SpecGroup:    specmodel.NewSpecGroup(client),
		SpecItem:     specmodel.NewSpecItem(client),
		Table:        table.NewTable(client),
		TableType:    table.NewTableType(client),
		Order:        order.NewOrder(client),
		Object:       objmodel.NewObject(client),
		Audit:        auditmodel.NewAudit(client),
		Settlement:   settlement.NewSettlement(client),
	}
	ctx.CasbinMiddleware = NewCasbinMiddleware(ctx)
	if err := SyncRolePoliciesToCasbin(ctx.Rbac, ctx.Casbin); err != nil {
		return nil, err
	}
	return ctx, nil
}
