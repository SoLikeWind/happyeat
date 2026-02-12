// 数据库迁移：根据 dal/model/ent 的 schema 创建/更新表结构。
// 使用方式：在项目根目录执行 go run ./app/cmd/migrate -f app/etc/happyeatservice.yaml，或 make migrate。
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"

	"github.com/solikewind/happyeat/app/internal/config"
	"github.com/solikewind/happyeat/dal/model/ent"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/happyeatservice.yaml", "config file path (relative to app dir)")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	db, err := sql.Open("pgx", c.SqlConfig.DataSource)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("migrate create: %v", err)
	}

	log.Println("migration done")
}
