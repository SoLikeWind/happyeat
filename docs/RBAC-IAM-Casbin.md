# RBAC / IAM / Casbin 说明与运维

本文档记录 HappyEat 权限体系的设计、2026-07 前后发生的 403 事故原因、已知隐患，以及日常运维习惯。**IAM 为业务真相源，Casbin 为 API 鉴权投影。**

---

## 1. 架构概览

```
登录 JWT
  └─ sub / user_code  ──► Casbin 中间件 Enforce(user_code, path, method)
                              │
                              ├─ g(user_code, role_code)   ← casbin_rule 表 (ptype=g)
                              └─ p(role_code, obj, act)    ← casbin_rule 表 (ptype=p)

IAM（PostgreSQL）
  ├─ iam_users / iam_user_roles     用户 ↔ 角色
  └─ iam_roles / iam_role_permissions  角色 ↔ 权限码

权限码 ──► PermissionRules ──► (obj, act)  HTTP 资源点
```

| 组件 | 作用 |
|------|------|
| **IAM** | 用户、角色、权限码的增删改查；`/iam/me`、登录 JWT 中的 `roles` 均来自 IAM |
| **Casbin** | 除白名单外的所有 `/central/v1/*` API 鉴权 |
| **前端权限矩阵** | 菜单/按钮显隐（`happyeat-web` 本地 + 后端 RBAC 接口拉取） |

**重要：** JWT 里的 `roles` **不参与**后端 Casbin 校验。后端只认 `user_code`（JWT `sub`）在 Casbin 里是否有有效的 `g` 绑定，以及对应角色是否有 `p` 策略。

---

## 2. 自助接口白名单

以下路径 **仅需登录**，不做 Casbin 策略校验（见 `app/internal/pkg/routenorm/routenorm.go`）：

| 路径 | 说明 |
|------|------|
| `GET /central/v1/iam/me` | 当前用户资料 |
| `POST /central/v1/iam/me/password` | 修改自己的密码 |
| `POST /central/v1/auth/login` | 登录（公开） |

因此事故期间常见现象：**`/iam/me` 返回正常角色，其它业务接口全部 403**——说明 IAM 有绑定，Casbin 未同步或绑定丢失。

---

## 3. 2026-07 403 事故摘要

### 现象

- 除 `dev-admin`（`super_admin`）外，其余账号几乎全部 403
- `GET /iam/me` 仍能看到正确 `roles`

### 根因

1. **IAM 与 `casbin_rule` 表不同步（删 unknown 触发的典型场景）**  
   历史用户常见状态：
   - **IAM**：`cashier` + `unknown`（或 `super_admin` + `unknown`）
   - **Casbin**：只有 `g(user, unknown)`，**从未写入** `g(user, cashier)`  
   旧版本可能还曾给 `unknown` 投影过 `p` 策略，用户靠 `unknown` 在 Casbin 里「蹭」权限。  
   在用户管理里**去掉 unknown** 时，旧逻辑只执行 `RemoveGroupingPolicy(user, unknown)`，**不会**把 IAM 里已有的真实角色补进 Casbin → 该账号立刻 403。  
   换另一个还有 `unknown` 的账号仍正常；**任意账号**只要删掉 unknown，都会复现——与具体账号无关。

2. **`unknown` 角色的历史错误**  
   - 本意：前端 `normalizeRole()` 的兜底码，**0 权限，不应写入 IAM**  
   - 历史：`defaultRolePermissions()` 曾含 `unknown`，bootstrap 将其 seed 进 `iam_roles`  
   - 部分用户仅绑定 `unknown`；删除该角色后 IAM/Casbin 均无有效角色 → 403  

3. **`init_policy.sql` 与 IAM 双来源**  
   手工 SQL 曾为 `dev-admin` 插入 `g(dev-admin, super_admin)`，在 Casbin 全量同步异常时，可能出现「只有 dev-admin 能用、其他人全挂」的假象。

### 恢复方式（已验证有效）

任选其一：

1. 超管登录 Web → **权限管理** → **同步 Casbin**（`POST /central/v1/rbac/casbin/sync`）
2. **重启后端**：`NewServiceContext` 启动时会执行 `SyncRolePoliciesToCasbin`

恢复后确认：各账号 Casbin `g` 与 IAM 一致，且无 `unknown` 绑定。

---

## 4. 预置角色与 dev-admin

| role_code | 中文名 | 说明 |
|-----------|--------|------|
| `super_admin` | 超级管理员 | 全部权限 |
| `manager` | 店长 | 全部权限（与超管矩阵相同，可再细分） |
| `cashier` | 收银 | 订单、点餐台、统计、结账单等 |
| `kitchen` | 后厨 | 工作台、订单等 |
| `waiter` | 服务员 | 点餐台、餐桌、订单等 |

**开发内置账号：**

| 字段 | 值 |
|------|-----|
| user_code | `dev-admin` |
| 登录用户名 | `admin`（映射到 `dev-admin`） |
| 默认密码 | `admin123` |
| 角色 | `super_admin` |

预置角色与 `unknown` 为**系统保护角色**，不可在 UI 删除（见 `casbinrules.IsProtectedRole`）。

---

## 5. 同步机制（何时写 Casbin）

| 操作 | IAM | Casbin |
|------|-----|--------|
| 服务启动 | bootstrap | `SyncRolePoliciesToCasbin` 全量 |
| 分配用户角色（API） | ✓ | `EnsureUserCasbinGroupings` 按 IAM 对齐 g |
| 移除用户角色（API） | ✓ | `EnsureUserCasbinGroupings` 按 IAM 对齐 g |
| 删除角色（API） | ✓ | 增量 `RemoveRolePoliciesFromCasbin` |
| 更新角色权限 / 重置权限 | ✓ | 全量 `SyncRolePoliciesToCasbin` |
| 删除用户 | ✓ | 全量 `SyncRolePoliciesToCasbin` |
| bootstrap 给 dev-admin 绑角色 | ✓ | **不单独写**（依赖启动全量同步） |

相关代码：

- `app/internal/svc/rbac_sync.go`
- `app/internal/svc/rbac_store.go`
- `app/internal/svc/servicecontext.go`（启动同步）

---

## 6. 日常运维习惯（推荐）

**在以下操作之后，执行一次「同步 Casbin」或重启后端：**

- 为用户分配 / 移除角色
- 修改角色权限矩阵
- 删除角色
- 删除用户
- 数据库手工改过 IAM 相关表

**排查 403 步骤：**

1. 用该账号调 `GET /iam/me`，看 `roles` 是否为空  
   - 为空 → 在用户管理重新分配角色，并重新登录  
2. `roles` 正常仍 403 → **IAM 与 Casbin 不同步**  
   - 超管执行 Casbin 同步或重启服务  
3. 仍异常 → 在数据库对比 IAM 与 Casbin（见第 8 节 SQL）

**请勿：**

- 生产环境手改 `casbin_rule` 表（除非紧急救火且事后立刻全量同步）
- 删除预置角色或再次引入 `unknown` 到 IAM

---

## 7. 已知隐患（暂未改代码，需知晓）

以下问题在 2026-07 审查中确认，**当前靠运维流程兜底**：

### P0 — 同步可靠性

- **双数据源**：IAM 与 Casbin 可能短暂或不一致；`/iam/me` 正常 ≠ API 有权限  
- **删 unknown 曾导致 403**（已修复）：解除 unknown 后须按 IAM 补齐真实角色的 Casbin `g`；见 `EnsureUserCasbinGroupings`  
- **全量同步 `SyncRolePoliciesToCasbin`**：曾观测到丢失 `dev-admin` 的 `g` 绑定、无法清理历史孤儿策略；`AddGroupingPolicy` 返回 `false` 时无日志  
- **`init_policy.sql`**：与 IAM 投影并存，dev-admin 可能被手工 SQL「特殊照顾」

### P1 — 数据与操作

- **非空库启动**时 `seedDefaultMappings` 只强制刷新 `super_admin` 权限，其它预置角色 permissions 若被改坏不会自动修复  
- **移除用户最后一个角色**目前无拦截，可能导致零角色 → 全站 403  
- **`ListUserRoles` 不过滤 `unknown`**：若 IAM 仍残留 `unknown` 绑定，会投影为无权限的 `g`

### P2 — 体验

- **`PUT /central/v1/iam/me`** 不在自助白名单，也无对应 `PermissionRules` → 非超管类角色改昵称/头像可能 403  
- **`home:view`** 权限码无后端 HTTP 映射，仅前端导航使用  
- Casbin 拒绝返回 **403**，前端仅在 **401** 清 token → 用户易误判为「登录坏了」

若后续要改代码，建议优先：**单一真相源（去掉手工 SQL 依赖）、加固全量同步、bootstrap 增量写 Casbin、启动时修复全部预置角色权限**。

---

## 8. 数据库对比（深度排查）

确认 IAM 用户–角色绑定：

```sql
SELECT u.user_code, r.role_code
FROM iam_user_roles ur
JOIN iam_users u ON u.id = ur.iam_user_id
JOIN iam_roles r ON r.id = ur.iam_role_id
WHERE u.delete_ts = 0 AND r.delete_ts = 0
ORDER BY u.user_code, r.role_code;
```

确认 Casbin 分组策略（列名为 `p_type`，非 `ptype`）：

```sql
SELECT v0 AS user_code, v1 AS role_code
FROM casbin_rule
WHERE p_type = 'g'
ORDER BY v0, v1;
```

两边 `user_code → role_code` 应一致。若 IAM 有而 Casbin 无，执行「同步 Casbin」或重启服务。

---

## 9. 关键文件索引

| 区域 | 路径 |
|------|------|
| RBAC 核心 / bootstrap | `app/internal/svc/rbac_store.go` |
| Casbin 同步 | `app/internal/svc/rbac_sync.go` |
| 启动注入 | `app/internal/svc/servicecontext.go` |
| Casbin 中间件 | `app/internal/svc/casbin_middleware.go` |
| 权限码与路由映射 | `app/internal/pkg/casbinrules/rules.go` |
| 路径归一化 / 白名单 | `app/internal/pkg/routenorm/routenorm.go` |
| IAM DAL | `dal/model/iam/iam.go` |
| Casbin 表结构示例 | `dal/casbin/init_policy.sql` |
| Web 权限页 / 同步按钮 | `happyeat-web/src/pages/PermissionManage.tsx` |
| Web 鉴权上下文 | `happyeat-web/src/contexts/AuthContext.tsx` |

---

## 10. 配置参考

默认数据库（见 `etc/happyeatservice.yaml`）：

```
postgres://postgres:123456@127.0.0.1:5432/happyeat
```

Local 配置可能指向 `happyeat_remote` 等其它库；**诊断与线上务必使用同一配置文件**，避免「IAM 看着正常、实际连的是另一个库」。

---

*文档版本：2026-07-03，对应 happyeat IAM/Casbin 重构及 unknown 清理之后的状态。*
