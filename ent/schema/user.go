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
			Unique(),
		field.String("first_name").NotEmpty(),
		field.String("last_name").NotEmpty(),
		field.Enum("gender").Values("other", "female", "male").Default("other"),
		field.String("phone").Unique().NotEmpty(),
		field.String("email").Unique().Optional().Nillable(),
		field.String("avatar").Optional().Nillable(),
		field.String("ward_code").Optional().Nillable(),
		field.String("address").Optional().Nillable(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("account", Account.Type).Unique(),
	}
}
