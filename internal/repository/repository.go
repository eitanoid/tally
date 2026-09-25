package repository

import (
	"context"

	"github.com/eitanoid/habit-tracker/internal/entries"
	"github.com/eitanoid/habit-tracker/internal/schemas"
)

type Repository interface {
	// Schema Operations
	InsertSchema(ctx context.Context, schema *schemas.HabitSchema) error
	GetLatestSchemaByID(ctx context.Context, habitID string) (*schemas.HabitSchema, error)
	GetSchemaByRef(ctx context.Context, ref schemas.SchemaRef) (*schemas.HabitSchema, error)
	GetAllLatestSchemas(ctx context.Context) ([]schemas.HabitSchema, error)

	// Entry Operations
	InsertEntry(ctx context.Context, entry *entries.HabitEntry) error
	GetEntriesByHabitID(ctx context.Context, habitID string, limit int) ([]entries.HabitEntry, error)
}
