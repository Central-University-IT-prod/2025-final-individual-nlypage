package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Targeting holds the schema definition for the Targeting entity.
type Targeting struct {
	ent.Schema
}

// Fields of the Targeting.
func (Targeting) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("gender").
			Values("MALE", "FEMALE", "ALL").
			Optional().
			Nillable(),
		field.Int("age_from").
			Optional().
			Nillable(),
		field.Int("age_to").
			Optional().
			Nillable(),
		field.String("location").
			Optional().
			Nillable(),
	}
}

// Edges of the Targeting.
func (Targeting) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).
			Ref("targeting").
			Required().
			Unique(),
	}
}

// Indexes of the Targeting.
func (Targeting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("location").
			Edges("campaign"),
		index.Fields("age_from", "age_to").
			Edges("campaign"),
	}
}
