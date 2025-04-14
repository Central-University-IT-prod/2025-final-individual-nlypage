package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Campaign holds the schema definition for the Campaign entity.
type Campaign struct {
	ent.Schema
}

// Fields of the Campaign.
func (Campaign) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique(),
		field.UUID("advertiser_id", uuid.UUID{}),
		field.Int("impressions_limit").
			Positive(),
		field.Int("clicks_limit").
			Positive(),
		field.Float("cost_per_impression"),
		field.Float("cost_per_click"),
		field.String("ad_title").
			NotEmpty(),
		field.String("ad_text").
			NotEmpty(),
		field.String("image_url").
			Optional(),
		field.Int("start_date").
			NonNegative(),
		field.Int("end_date").
			NonNegative(),
		field.Bool("moderated"),
	}
}

// Edges of the Campaign.
func (Campaign) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("targeting", Targeting.Type).
			Unique(),
	}
}

// Indexes of the Campaign.
func (Campaign) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("start_date", "end_date"),
	}
}
