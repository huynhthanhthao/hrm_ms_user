package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Unique().
			StructTag(`json:"id"`),
		field.String("first_name").
			NotEmpty().
			StructTag(`json:"first_name"`),
		field.String("last_name").
			NotEmpty().
			StructTag(`json:"last_name"`),
		field.Enum("gender").
			Values("other", "female", "male").
			Default("other").
			StructTag(`json:"gender"`),
		field.String("phone").
			Unique().
			NotEmpty().
			StructTag(`json:"phone"`),
		field.String("email").
			Unique().
			Optional().
			Nillable().
			StructTag(`json:"email"`),
		field.String("avatar").
			Optional().
			Nillable().
			StructTag(`json:"avatar"`),
		field.String("ward_code").
			Optional().
			Nillable().
			StructTag(`json:"ward_code"`),
		field.String("address").
			Optional().
			Nillable().
			StructTag(`json:"address"`),
		field.Time("created_at").
			Default(time.Now).
			StructTag(`json:"created_at"`),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			StructTag(`json:"updated_at"`),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("account", Account.Type).Unique(),
	}
}
