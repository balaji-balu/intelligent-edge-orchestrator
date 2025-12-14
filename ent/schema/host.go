package schema

import (
    "entgo.io/ent"
	//"entgo.io/ent/schema"
    //"entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    //"entgo.io/ent/dialect/entsql"
    "github.com/google/uuid"
)

type Host struct {
    ent.Schema
}

// Primary key
// func (Host) ID() ent.Field {
//     return field.UUID("host_id", uuid.UUID{}).Immutable()
// }

func (Host) Fields() []ent.Field {
    return []ent.Field{
		//UUIDField(),
		field.UUID("host_id", uuid.UUID{}).Unique(),
        // Foreign key to Site
        field.UUID("site_id", uuid.UUID{}), //Optional().Nillable(),

        field.String("runtime").Optional(),
        field.Time("last_seen").Optional(),
        field.Int("misses").Optional(),
        field.Float("cpu_free").Optional(),
        field.String("status").Optional(),
        field.String("hostname").Optional(),
        field.String("ip_address").Optional(),
        field.String("edge_url").Optional(),
        field.JSON("metadata", map[string]any{}).Optional(),
        field.Time("created_at").Optional(),
        field.Time("updated_at").Optional(),
    }
}

// func (Host) Edges() []ent.Edge {
//     return []ent.Edge{
//         edge.From("site", Site.Type).   // the parent entity
//             Ref("hosts").               // reverse edge on Site
//             Field("site_id").           // FK field in Host
//             Required(),                 // not nullable
//     }
// }

