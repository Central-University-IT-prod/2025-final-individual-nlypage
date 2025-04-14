package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Advertiser holds the schema definition for the Advertiser entity.
type Advertiser struct {
	ent.Schema
}

// Fields of the Advertiser.
func (Advertiser) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique(),
		field.String("name").
			NotEmpty(),
	}
}

// Edges of the Advertiser.
func (Advertiser) Edges() []ent.Edge {
	return nil
}
