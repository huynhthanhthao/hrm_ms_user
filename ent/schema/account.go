package schema

import (
	"time"

	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Account holds the schema definition for the Account entity.
type Account struct {
	ent.Schema
}

// Fields of the Account.
func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Unique().NotEmpty().
			Annotations(
				entproto.Field(2),
			),
		field.String("password").Sensitive().NotEmpty().
			Annotations(
				entproto.Field(3),
			),
		field.Enum("status").
			Values("active", "inactive").
			Default("active").
			Annotations(
				entproto.Field(4),
				entproto.Enum(map[string]int32{
					"active":   0,
					"inactive": 1,
				}),
			),
		field.Time("created_at").Default(time.Now).
			Annotations(
				entproto.Field(5),
			),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).
			Annotations(
				entproto.Field(6),
			),
		field.Int("user_id").
			Positive().
			Annotations(
				entproto.Field(8),
			),
	}
}

// Edges of the Account.
func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("account").
			Unique().
			Required().
			Field("user_id").
			Annotations(entproto.Field(7)),
	}
}

func (Account) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entproto.Message(),
		entproto.Service(),
	}
}
