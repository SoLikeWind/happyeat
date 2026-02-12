.PHONY: api migrate run swagger swagger-serve
# 统一入口：一次生成 menu+table 的 types、handler、logic，避免相互覆盖
api:
	goctl api go --api app/api/v1/central.api --dir app

# 数据库迁移（在 app 目录执行，需先配置 app/etc/happyeatservice.yaml 的 DataSource）
migrate:
	cd app && go run ./cmd/migrate -f etc/happyeatservice.yaml

# 启动 HTTP 服务（需先执行 make migrate，可选执行 dal/casbin/init_policy.sql）
run:
	cd app && go run . -f etc/happyeatservice.yaml

# 根据 central.api 生成 Swagger 接口文档（JSON 默认生成到项目根目录 happyeat.json；需 goctl >= 1.8.2）
swagger:
	goctl api swagger --api app/api/v1/central.api --dir . --filename happyeat

# 启动静态服务供 Swagger 预览：执行后浏览器打开 https://editor.swagger.io 并填入 http://localhost:3780/happyeat.json
swagger-serve:
	npx --yes http-server . -p 3780 -c-1 --cors 