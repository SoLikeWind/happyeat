-- Casbin 策略表：与 pckhoi/casbin-pgx-adapter 建表结构一致（id 为 text 主键，列为 p_type 与 v0~v5）。
-- 若使用 adapter 启动，表会自动创建；也可本脚本手动建表并插入示例数据。

CREATE TABLE IF NOT EXISTS casbin_rule (
  id     text PRIMARY KEY,
  p_type text,
  v0     text,
  v1     text,
  v2     text,
  v3     text,
  v4     text,
  v5     text
);

-- 示例：角色 admin 对任意资源 * 拥有任意操作 *；用户 alice 属于角色 admin（按需执行，避免重复插入）
-- id 可与 adapter 一致：在项目根执行 go run ./app/cmd/genpolicyid p admin '*' '*' 得到 p 策略的 id，下同
INSERT INTO casbin_rule (id, p_type, v0, v1, v2) VALUES ('p-admin-wildcard', 'p', 'admin', '*', '*');
INSERT INTO casbin_rule (id, p_type, v0, v1) VALUES ('g-alice-admin', 'g', 'alice', 'admin');
