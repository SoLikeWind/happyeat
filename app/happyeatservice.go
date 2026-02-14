// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/solikewind/happyeat/app/internal/config"
	"github.com/solikewind/happyeat/app/internal/handler"
	"github.com/solikewind/happyeat/app/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/happyeatservice.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCustomCors(
		func(h http.Header) {
			h.Set("Access-Control-Allow-Origin", "*")
			h.Add("Access-Control-Allow-Headers", "Content-Type, Authorization")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			h.Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		},
		func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "CORS 拒绝"}`))
		}, "*"))
	defer server.Stop()

	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.DB.Close()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
