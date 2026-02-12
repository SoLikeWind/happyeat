-- Casbin 策略表：若使用 pgx-adapter，首次连接时会自动建表；也可手动执行本脚本建表并插入示例数据。
-- 与 app/etc/happyeatservice.yaml 中 casbin.model 的 RBAC 模型对应：p = sub,obj,act；g = 用户,角色。

CREATE TABLE IF NOT EXISTS casbin_rule (
  id    BIGSERIAL PRIMARY KEY,
  ptype VARCHAR(100) NOT NULL DEFAULT '',
  v0    VARCHAR(100) NOT NULL DEFAULT '',
  v1    VARCHAR(100) NOT NULL DEFAULT '',
  v2    VARCHAR(100) NOT NULL DEFAULT '',
  v3    VARCHAR(100) NOT NULL DEFAULT '',
  v4    VARCHAR(100) NOT NULL DEFAULT '',
  v5    VARCHAR(100) NOT NULL DEFAULT ''
);

-- 示例：角色 admin 对任意资源 * 拥有任意操作 *；用户 alice 属于角色 admin（按需执行，避免重复插入）
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'admin', '*', '*');
INSERT INTO casbin_rule (ptype, v0, v1) VALUES ('g', 'alice', 'admin');
