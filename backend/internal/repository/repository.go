// Package repository defines the data persistence interfaces and filters for tally domains.
package repository

import (
	"context"

	"github.com/eitanoid/tally/internal/entries"
	"github.com/eitanoid/tally/internal/schemas"
)

// EntryFilter contains the pagination filter for Entry queries.
type EntryFilter struct {
	TallyID string
	Limit   int
	Offset  int
}

// Repository contains the signatures for all database operations.
type Repository interface {
	// Schema Operations
	InsertSchema(ctx context.Context, schema *schemas.TallySchema) error
	GetLatestSchemaByID(ctx context.Context, tallyID string) (*schemas.TallySchema, error)
	GetSchemaByRef(ctx context.Context, ref schemas.SchemaRef) (*schemas.TallySchema, error)
	GetAllLatestSchemas(ctx context.Context) ([]schemas.TallySchema, error)

	// Entry Operations
	InsertEntry(ctx context.Context, entry *entries.TallyEntry) error
	GetEntriesByTallyID(ctx context.Context, filter EntryFilter) ([]entries.TallyEntry, int, error)
	GetEntryByID(ctx context.Context, entryID string) (*entries.TallyEntry, error)
	UpdateEntryData(ctx context.Context, entry *entries.TallyEntry) error

	Close() error
}
