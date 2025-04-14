package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MlScore holds the schema definition for the MlScore entity.
type MlScore struct {
	ent.Schema
}

// Fields of the MlScore.
func (MlScore) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			StorageKey("user_id").
			Immutable(),
		field.UUID("advertiser_id", uuid.UUID{}).
			StorageKey("advertiser_id").
			Immutable(),
		field.Int64("score"),
	}
}

// Indexes returns the indexes of the schema.
func (MlScore) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "advertiser_id").
			Unique(),
	}
}

// Edges of the MlScore.
func (MlScore) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id").
			Immutable(),
		edge.To("advertiser", Advertiser.Type).
			Unique().
			Required().
			Field("advertiser_id").
			Immutable(),
	}
}
