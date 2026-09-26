package repository

import (
	"context"

	"github.com/eitanoid/tally/internal/entries"
	"github.com/eitanoid/tally/internal/schemas"
)

type Repository interface {
	// Schema Operations
	InsertSchema(ctx context.Context, schema *schemas.TallySchema) error
	GetLatestSchemaByID(ctx context.Context, tallyID string) (*schemas.TallySchema, error)
	GetSchemaByRef(ctx context.Context, ref schemas.SchemaRef) (*schemas.TallySchema, error)
	GetAllLatestSchemas(ctx context.Context) ([]schemas.TallySchema, error)

	// Entry Operations
	InsertEntry(ctx context.Context, entry *entries.TallyEntry) error
	GetEntriesByTallyID(ctx context.Context, tallyID string, limit int) ([]entries.TallyEntry, error)

	// Search tallys by name
}
