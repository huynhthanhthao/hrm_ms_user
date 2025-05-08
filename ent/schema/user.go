package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("first_name").NotEmpty(),
		field.String("last_name").NotEmpty(),
		field.Enum("gender").Values("other", "female", "male").Default("other"),
		field.String("email").Unique().NotEmpty(),
		field.String("phone").Unique().NotEmpty(),
		field.String("ward_code").NotEmpty(),
		field.String("address").NotEmpty(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.String("company_id").NotEmpty(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("account", Account.Type).Unique(),
	}
}
