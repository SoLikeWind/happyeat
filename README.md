# HappyEat 后端

HappyEat 是一个面向餐饮门店的点餐与订单管理系统。当前仓库为后端服务，提供菜单、规格、餐桌、点餐、订单、结账单、厨房打印、权限等 API。

配套项目：

- 管理后台：`../happyeat-web`
- 移动端 App：`../happyeat-app`

## 项目状态

基本完结。核心点餐、订单、菜单、餐桌、结账单、厨房打印等流程已具备基础能力，仍在持续完善体验和稳定性。

## 技术栈

- Go
- go-zero：HTTP API 服务框架
- ent：ORM 与数据库模型
- Casbin：权限控制
- PostgreSQL：推荐生产数据库

## 快速开始

### 1. 准备配置

默认配置位于：

```text
app/etc/happyeatservice.yaml
```

如需生产环境覆盖，可参考：

```text
app/etc/happyeatservice.remote.yaml.example
```

### 2. 数据库迁移

```bash
make migrate
```

### 3. 启动服务

```bash
make run
```

服务默认监听：

```text
http://0.0.0.0:8888
```

接口前缀：

```text
/central/v1
```

## 常用开发命令

```bash
# 运行测试
go test ./...

# 启动后端
make run

# 执行数据库迁移
make migrate
```

如果当前环境首次运行测试，Go 可能会先下载 toolchain 和依赖，耗时会比平时更久。

## 主要模块

- `app/`：go-zero API 服务入口、路由、handler、logic、配置
- `dal/model/`：ent 生成代码与业务数据访问封装
- `common/`：通用工具
- `docs/`：设计文档与实施计划

## 相关文档

- 实施计划：`docs/implementation-plan.md`
- 厨房打印单格式设计：`docs/厨房打印单格式设计.md`
- RBAC / IAM / Casbin 运维与已知问题：`docs/RBAC-IAM-Casbin.md`
- API 说明：`app/api/README.md`

## 架构说明

当前后端采用单体模块化方式组织，API 层直接调用内部 logic 与数据访问层。这个形态更适合当前门店级业务：部署简单、联调成本低、问题定位快。

后续如果并发、团队协作或多服务边界需求变强，可以再拆分 RPC 服务或引入 gateway。拆分前需要权衡调用链、部署复杂度、监控链路和本地联调成本。