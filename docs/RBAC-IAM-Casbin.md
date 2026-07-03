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

4. **角色 CRUD 触发全量 Casbin 重建（权限页）**  
   旧版在创建 / 保存 / 重置角色权限时调用 `SyncRolePoliciesToCasbin`，会「先删光再重建」全部用户的 `g` 绑定；`AddGroupingPolicy` 失败时无感知 → **操作者本人**在刷新权限矩阵时立刻 403（`GET /rbac/role-permissions`、`GET /iam/roles`）。  
   **已修复**（`050e0d4`）：改为 `ReplaceRolePoliciesInCasbin` + `EnsureActorCasbinGroupings`。

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
| 服务启动 | bootstrap | `SyncRolePoliciesToCasbin`（`p` 全量 + 逐用户 `EnsureUserCasbinGroupings`） |
| 分配 / 移除用户角色（API） | ✓ | `EnsureUserCasbinGroupings` 按 IAM 对齐 g |
| 创建角色（带初始权限） | ✓ | `ReplaceRolePoliciesInCasbin` 仅写该角色 `p` + `EnsureActorCasbinGroupings` |
| 更新 / 重置角色权限 | ✓ | `ReplaceRolePoliciesInCasbin`（逐角色）+ `EnsureActorCasbinGroupings` |
| 删除角色（API） | ✓ | `RemoveRolePoliciesFromCasbin` + 受影响用户 `EnsureUserCasbinGroupings` + 操作者对齐 |
| 删除用户（API） | ✓ | `RemoveAllUserCasbinGroupings`（仅删该用户 g，不重建全员） |
| 手动「同步 Casbin」 | — | `SyncRolePoliciesToCasbin` |
| bootstrap 给 dev-admin 绑角色 | ✓ | 依赖启动时全量同步 |

核心函数（`app/internal/svc/rbac_sync.go`）：

| 函数 | 作用 |
|------|------|
| `EnsureUserCasbinGroupings` | 单用户：IAM → Casbin `g` 对齐，跳过 `unknown` |
| `EnsureUserCasbinGroupingsForUsers` | 批量用户对齐 |
| `EnsureActorCasbinGroupings` | 当前登录用户对齐（角色变更后防操作者 403） |
| `ReplaceRolePoliciesInCasbin` | 单角色 `p` 策略替换，**不碰**用户 `g` |
| `RemoveRolePoliciesFromCasbin` | 删角色：移除该角色 `p` + 相关 `g` |
| `RemoveAllUserCasbinGroupings` | 删用户：移除该用户全部 `g` |
| `SyncRolePoliciesToCasbin` | 全量：重建全部 `p` + 逐用户对齐 `g` |

相关代码：

- `app/internal/svc/rbac_sync.go`
- `app/internal/svc/rbac_store.go`
- `app/internal/svc/servicecontext.go`（启动同步）

---

## 6. 日常运维习惯（推荐）

**在以下操作之后，一般无需再手动同步**（代码已增量对齐）；若仍 403 可重启或点「同步 Casbin」：

- 为用户分配 / 移除角色
- 在权限页创建 / 保存 / 重置 / 删除角色
- 删除用户

**仍建议手动同步或重启的场景：**

- 数据库手工改过 IAM / `casbin_rule` 表
- 从旧版本升级后首次部署
- 排查 403 仍无法恢复时

**排查 403 步骤：**

1. 用该账号调 `GET /iam/me`，看 `roles` 是否为空  
   - 为空 → 在用户管理重新分配角色，并重新登录  
2. `roles` 正常仍 403 → **IAM 与 Casbin 不同步**  
   - 超管执行 Casbin 同步或重启服务  
3. 仍异常 → 在数据库对比 IAM 与 Casbin（见第 9 节 SQL）

**请勿：**

- 生产环境手改 `casbin_rule` 表（除非紧急救火且事后立刻全量同步）
- 删除预置角色或再次引入 `unknown` 到 IAM

---

## 7. Bug 与问题记录

### 7.1 已修复（2026-07-03，提交 `24b1b2f` / `050e0d4`）

| # | 问题 | 现象 | 根因 | 修复 |
|---|------|------|------|------|
| B1 | `unknown` 写入 IAM | 部分用户仅绑 `unknown`；删角色后全站 403 | bootstrap 曾 seed `unknown`；前端兜底码不应进 IAM | 不再 seed；启动 `purgeLegacyUnknownBindings`；UI/API 隐藏 |
| B2 | 删用户 `unknown` 后 403 | **任意账号**去掉 unknown 即 403，换号仍正常 | IAM 有真实角色，Casbin 仅 `g(user,unknown)`；旧逻辑只删 unknown 不补真实 `g` | `EnsureUserCasbinGroupings`（分配/移除用户角色时） |
| B3 | 全站 403（除 dev-admin） | `/iam/me` 有角色，其它 API 403 | IAM 与 `casbin_rule` 不同步；`init_policy.sql` 手工 g 与全量同步异常 | 启动/手动全量同步；文档化运维流程 |
| B4 | 角色 CRUD 后操作者 403 | 权限页创建/保存/重置角色后「加载权限数据失败」 | 角色权限变更触发 `SyncRolePoliciesToCasbin` 冲掉操作者 `g` | `ReplaceRolePoliciesInCasbin` + `EnsureActorCasbinGroupings` |
| B5 | 删角色 / 删用户误伤全员 | 删除后无关用户权限异常 | 删用户曾触发全量 Casbin 重建 | 删用户改 `RemoveAllUserCasbinGroupings`；删角色后补全受影响用户 `g` |
| B6 | 全量同步 `g` 不可靠 | 同步后丢 dev-admin 绑定、遗留孤儿 `g` | 「先 Remove 全部 g 再 Add」+ adapter 静默失败 | 全量同步的 `g` 部分改为逐用户 `EnsureUserCasbinGroupings` |
| B7 | 预置角色展示英文 | 员工看到 `cashier` 等 code | DB `role_name` 未回填 | `PresetRoleNames` + API `role_names` + 前端 `resolveRoleDisplayName` |

### 7.2 仍待观察 / 未修复

| # | 问题 | 影响 | 临时规避 |
|---|------|------|----------|
| O1 | `seedDefaultMappings` 非空库只刷新 `super_admin` | 其它预置角色 permissions 被改坏后不自动修复 | 权限页对该角色「重置默认」+ 同步 Casbin |
| O2 | 可移除用户最后一个角色 | 用户零角色 → 无法登录 / 403 | 删角色前确认用户另有角色 |
| O3 | `PUT /iam/me` 无 Casbin 规则 | 非超管改自己昵称/头像可能 403 | 管理员在用户管理页修改 |
| O4 | `home:view` 无后端 HTTP 映射 | 仅前端导航用，后端不因该码放行 | 无（设计如此） |
| O5 | Casbin 403 vs JWT 401 语义 | 用户以为「登录坏了」 | 看 `/iam/me` 是否有角色；有则同步 Casbin |
| O6 | `init_policy.sql` 与 IAM 双来源 | dev-admin 可能靠手工 SQL 撑权限 | 以 IAM + 同步为准，勿手改 `casbin_rule` |
| O7 | `AddGroupingPolicy` 返回 false 无日志 | 极端情况下绑定写入失败难排查 | 重启 + 全量同步；对比第 9 节 SQL |

若后续继续改代码，建议优先：**去掉 `init_policy.sql` 手工依赖**、**启动时修复全部预置角色 permissions**、**禁止移除用户最后一个角色**。

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

*文档版本：2026-07-03（`24b1b2f` unknown 清理与 IAM 同步；`050e0d4` 角色 CRUD 增量 Casbin）。*
