package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Car holds the schema definition for the Car entity.
type Car struct {
	ent.Schema
}

// Fields of the Car.
func (Car) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name").Optional(),
		field.String("VIN").Immutable().Unique(),
		field.Time("last_accessed").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Car.
func (Car) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("value_histories", ValueHistory.Type),
	}
}
