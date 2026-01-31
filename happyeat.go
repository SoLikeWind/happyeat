// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

	"happyeat/internal/config"
	"happyeat/internal/handler"
	"happyeat/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/happyeat-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c) // 加载配置

	server := rest.MustNewServer(c.RestConf) // 创建服务
	defer server.Stop()

	ctx := svc.NewServiceContext(c)       // 初始化服务上下文（通常包含中间件、数据库连接、Redis 客户端、日志等依赖）。
	handler.RegisterHandlers(server, ctx) // 注册路由

	logConf := logx.LogConf{
		ServiceName: c.Log.ServiceName,
		Mode:        c.Log.Mode,
		Encoding:    c.Log.Encoding,
		Path:        c.Log.Path,
		Filename:    c.Log.Filename,
		MaxSize:     c.Log.MaxSize,
		MaxAge:      c.Log.MaxAge,
		MaxBackups:  c.Log.MaxBackups,
	}
	logx.MustSetup(logConf) // 设置日志

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
