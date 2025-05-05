package schema

import (
	"time"

	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("first_name").NotEmpty().
			Annotations(
                entproto.Field(2),
            ),
		field.String("last_name").NotEmpty().
			Annotations(
                entproto.Field(3),
            ),
		field.Enum("gender").
			Values("other", "female", "male").
			Default("other").
			Annotations(
                entproto.Field(4),
				entproto.Enum(map[string]int32{
					"other": 0,
					"female": 1,
					"male":  2,
				}),
            ),
		field.String("email").Unique().NotEmpty().
			Annotations(
                entproto.Field(5),
            ),
		field.String("phone").Unique().NotEmpty().
			Annotations(
                entproto.Field(6),
            ),
		field.String("ward_code").NotEmpty().
			Annotations(
                entproto.Field(7),
            ),
		field.String("address").NotEmpty().
			Annotations(
                entproto.Field(8),
            ),
		field.Time("created_at").Default(time.Now).
			Annotations(
                entproto.Field(9),
            ),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).
			Annotations(
                entproto.Field(10),
            ),
		field.String("company_id").NotEmpty().
			Annotations(
                entproto.Field(12),
            ),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("account", Account.Type).Unique().
			Annotations(entproto.Field(11)),
	}
}

func (User) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entproto.Message(),
        entproto.Service(),
    }
}

