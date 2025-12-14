package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/edge"
	"github.com/google/uuid"
)
type Site struct {
    ent.Schema
}

func (Site) Fields() []ent.Field {
    return []ent.Field{
		//UUIDField(),
		field.UUID("site_id", uuid.UUID{}).Unique(),
        field.String("name").Optional(),
        field.String("description").Optional(),
        field.String("location").Optional(),
        field.UUID("orchestrator_id", uuid.UUID{}).Optional().Nillable(),
        field.JSON("metadata", map[string]any{}).Optional(),
        field.Time("created_at").Optional(),
        field.Time("updated_at").Optional(),
    }
}

func (Site) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("hosts", Host.Type),
    }
}
