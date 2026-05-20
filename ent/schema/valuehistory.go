package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ValueHistory holds the schema definition for the ValueHistory entity.
type ValueHistory struct {
	ent.Schema
}

// Fields of the ValueHistory.
func (ValueHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Immutable().Default(time.Now),
		field.Enum("type").Values("READ", "WRITE").Immutable(),
		field.Bytes("UEC").Immutable(),
		field.Bytes("UECPWM").Immutable(),
		field.Bytes("REC").Immutable(),
	}
}

// Edges of the ValueHistory.
func (ValueHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("car", Car.Type).Ref("value_histories").Unique().Required(),
	}
}
