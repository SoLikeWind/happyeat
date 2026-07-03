package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// OperationLog 记录后台操作审计日志。
type OperationLog struct {
	ent.Schema
}

func (OperationLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UniqueID{},
		TimeMixin{},
	}
}

func (OperationLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "operation_logs"},
		schema.Comment("后台操作审计日志"),
	}
}

func (OperationLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("actor_user_code").
			MaxLen(128).
			Comment("操作者 user_code"),
		field.String("module").
			MaxLen(64).
			Comment("业务模块"),
		field.String("action").
			MaxLen(64).
			Comment("动作"),
		field.String("method").
			MaxLen(16).
			Comment("HTTP 方法"),
		field.String("path").
			MaxLen(512).
			Comment("原始请求路径"),
		field.String("normalized_path").
			MaxLen(512).
			Comment("归一化路径"),
		field.Int("status").
			Default(0).
			Comment("响应状态码"),
		field.String("target_id").
			MaxLen(128).
			Default("").
			Comment("目标 ID（从路径提取，可能为空）"),
		field.String("ip").
			MaxLen(128).
			Default("").
			Comment("客户端 IP"),
		field.String("user_agent").
			MaxLen(512).
			Default("").
			Comment("User-Agent"),
		field.String("summary").
			MaxLen(512).
			Default("").
			Comment("操作摘要"),
	}
}
