package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Account struct {
	ent.Schema
}

func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Unique().
			StructTag(`json:"id"`),
		field.String("username").
			Unique().
			NotEmpty().
			StructTag(`json:"username"`),
		field.String("password").
			Sensitive().
			NotEmpty(),
		field.Enum("status").
			Values("active", "inactive").
			Default("active").
			StructTag(`json:"status"`),
		field.Time("created_at").
			Default(time.Now).
			StructTag(`json:"created_at"`),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			StructTag(`json:"updated_at"`),
	}
}

func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("account").Unique().Required(),
	}
}
