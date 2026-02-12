// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	SqlConfig SqlConfig
	Auth      Auth
	Casbin    Casbin
}
type SqlConfig struct {
	DataSource string
}

type Auth struct {
	AccessSecret string
	AccessExpire int64
}
type Casbin struct {
	Model string // 模型内联字符串（与 yaml 中 casbin.model 一致）；策略从数据库 casbin_rule 表加载
}
