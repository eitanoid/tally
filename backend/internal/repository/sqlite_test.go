// Package repository_test verifies SQLite persistence layer operations using table-driven tests.
package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/eitanoid/tally/internal/entries"
	"github.com/eitanoid/tally/internal/repository"
	"github.com/eitanoid/tally/internal/schemas"
)

func setupTestDB(t *testing.T) *repository.SQLiteClient {
	t.Helper()

	client, err := repository.NewSQLiteClient(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	return client
}

func TestInsertSchema(t *testing.T) {
	tests := []struct {
		name    string
		schema  *schemas.TallySchema
		wantErr bool
	}{
		{
			name: "valid schema insertion",
			schema: &schemas.TallySchema{
				TallyID:       "550e8400-e29b-41d4-a716-446655440000",
				Version:       1,
				Name:          "Habit Tracker",
				Description:   "Daily habit tracking",
				JSONSchemaRaw: `{"type":"object"}`,
			},
			wantErr: false,
		},
		{
			name:    "nil schema error",
			schema:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := setupTestDB(t)
			ctx := context.Background()

			err := client.InsertSchema(ctx, tt.schema)
			if (err != nil) != tt.wantErr {
				t.Fatalf("InsertSchema() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.schema.CreatedAt.IsZero() {
				t.Errorf("expected CreatedAt to be set via RETURNING clause, got zero time")
			}
		})
	}
}

func TestGetLatestSchemaByID(t *testing.T) {
	tests := []struct {
		name        string
		seedSchemas []*schemas.TallySchema
		queryID     string
		wantVersion int
		wantErrIs   error
	}{
		{
			name: "returns latest version among multiple",
			seedSchemas: []*schemas.TallySchema{
				{TallyID: "550e8400-e29b-41d4-a716-446655440000", Version: 1, Name: "Tracker v1", JSONSchemaRaw: `{}`},
				{TallyID: "550e8400-e29b-41d4-a716-446655440000", Version: 2, Name: "Tracker v2", JSONSchemaRaw: `{}`},
			},
			queryID:     "550e8400-e29b-41d4-a716-446655440000",
			wantVersion: 2,
			wantErrIs:   nil,
		},
		{
			name:        "non-existent tally_id returns ErrNoRows",
			seedSchemas: nil,
			queryID:     "00000000-0000-0000-0000-000000000000",
			wantVersion: 0,
			wantErrIs:   sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := setupTestDB(t)
			ctx := context.Background()

			for _, s := range tt.seedSchemas {
				if err := client.InsertSchema(ctx, s); err != nil {
					t.Fatalf("failed to seed schema: %v", err)
				}
			}

			got, err := client.GetLatestSchemaByID(ctx, tt.queryID)
			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("GetLatestSchemaByID() error = %v, wantErrIs %v", err, tt.wantErrIs)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Version != tt.wantVersion {
				t.Errorf("GetLatestSchemaByID().Version = %d, want %d", got.Version, tt.wantVersion)
			}
		})
	}
}

func TestGetSchemaByRef(t *testing.T) {
	tests := []struct {
		name        string
		seedSchemas []*schemas.TallySchema
		ref         schemas.SchemaRef
		wantVersion int
		wantErrIs   error
	}{
		{
			name: "fetches exact version by ref",
			seedSchemas: []*schemas.TallySchema{
				{TallyID: "550e8400-e29b-41d4-a716-446655440000", Version: 1, Name: "v1", JSONSchemaRaw: `{}`},
				{TallyID: "550e8400-e29b-41d4-a716-446655440000", Version: 2, Name: "v2", JSONSchemaRaw: `{}`},
			},
			ref:         schemas.SchemaRef{TallyID: "550e8400-e29b-41d4-a716-446655440000", Version: 1},
			wantVersion: 1,
			wantErrIs:   nil,
		},
		{
			name:        "missing version returns ErrNoRows",
			seedSchemas: nil,
			ref:         schemas.SchemaRef{TallyID: "550e8400-e29b-41d4-a716-446655440000", Version: 99},
			wantVersion: 0,
			wantErrIs:   sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := setupTestDB(t)
			ctx := context.Background()

			for _, s := range tt.seedSchemas {
				if err := client.InsertSchema(ctx, s); err != nil {
					t.Fatalf("failed to seed schema: %v", err)
				}
			}

			got, err := client.GetSchemaByRef(ctx, tt.ref)
			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("GetSchemaByRef() error = %v, wantErrIs %v", err, tt.wantErrIs)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Version != tt.wantVersion {
				t.Errorf("GetSchemaByRef().Version = %d, want %d", got.Version, tt.wantVersion)
			}
		})
	}
}

func TestGetAllLatestSchemas(t *testing.T) {
	tests := []struct {
		name        string
		seedSchemas []*schemas.TallySchema
		wantIDs     []string
		wantErr     bool
	}{
		{
			name: "returns only latest versions sorted by name ASC",
			seedSchemas: []*schemas.TallySchema{
				{TallyID: "b-uuid", Version: 1, Name: "Coffee Log", JSONSchemaRaw: `{}`},
				{TallyID: "a-uuid", Version: 1, Name: "Book Tracker v1", JSONSchemaRaw: `{}`},
				{TallyID: "a-uuid", Version: 2, Name: "Book Tracker v2", JSONSchemaRaw: `{}`},
			},
			wantIDs: []string{"a-uuid", "b-uuid"}, // "Book Tracker v2" comes before "Coffee Log"
			wantErr: false,
		},
		{
			name:        "returns empty slice when table is empty",
			seedSchemas: nil,
			wantIDs:     []string{},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := setupTestDB(t)
			ctx := context.Background()

			for _, s := range tt.seedSchemas {
				if err := client.InsertSchema(ctx, s); err != nil {
					t.Fatalf("failed to seed schema: %v", err)
				}
			}

			got, err := client.GetAllLatestSchemas(ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAllLatestSchemas() error = %v, wantErr %v", err, tt.wantErr)
			}

			gotIDs := make([]string, 0, len(got))
			for _, s := range got {
				gotIDs = append(gotIDs, s.TallyID)
			}

			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Errorf("GetAllLatestSchemas() got TallyIDs = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestInsertEntry(t *testing.T) {
	tallyUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name    string
		entry   *entries.TallyEntry
		wantErr bool
	}{
		{
			name: "valid entry insertion",
			entry: &entries.TallyEntry{
				ID:            "entry-1",
				TallyID:       tallyUUID,
				SchemaVersion: 1,
				Data:          `{"value":42}`,
			},
			wantErr: false,
		},
		{
			name:    "nil entry error",
			entry:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := setupTestDB(t)
			ctx := context.Background()

			// Seed schema first for foreign key constraint
			if err := client.InsertSchema(ctx, &schemas.TallySchema{
				TallyID:       tallyUUID,
				Version:       1,
				Name:          "Test Tally",
				JSONSchemaRaw: `{}`,
			}); err != nil {
				t.Fatalf("failed to seed required schema: %v", err)
			}

			err := client.InsertEntry(ctx, tt.entry)
			if (err != nil) != tt.wantErr {
				t.Fatalf("InsertEntry() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && (tt.entry.CreatedAt.IsZero() || tt.entry.UpdatedAt.IsZero()) {
				t.Errorf("expected CreatedAt and UpdatedAt to be set")
			}
		})
	}
}

func TestGetEntriesByTallyID(t *testing.T) {
	tallyUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name        string
		seedEntries []*entries.TallyEntry
		filter      repository.EntryFilter
		wantCount   int
		wantTotal   int
		wantErr     bool
	}{
		{
			name: "fetches paginated entries and total count",
			seedEntries: []*entries.TallyEntry{
				{ID: "e1", TallyID: tallyUUID, SchemaVersion: 1, Data: `{}`},
				{ID: "e2", TallyID: tallyUUID, SchemaVersion: 1, Data: `{}`},
				{ID: "e3", TallyID: tallyUUID, SchemaVersion: 1, Data: `{}`},
			},
			filter: repository.EntryFilter{
				TallyID: tallyUUID,
				Limit:   2,
				Offset:  0,
			},
			wantCount: 2,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name: "handles negative limits and offsets gracefully",
			seedEntries: []*entries.TallyEntry{
				{ID: "e1", TallyID: tallyUUID, SchemaVersion: 1, Data: `{}`},
			},
			filter: repository.EntryFilter{
				TallyID: tallyUUID,
				Limit:   -10, // clamped to default 50
				Offset:  -5,  // clamped to 0
			},
			wantCount: 1,
			wantTotal: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := setupTestDB(t)
			ctx := context.Background()

			// Seed required schema
			if err := client.InsertSchema(ctx, &schemas.TallySchema{
				TallyID:       tallyUUID,
				Version:       1,
				Name:          "Test Tally",
				JSONSchemaRaw: `{}`,
			}); err != nil {
				t.Fatalf("failed to seed schema: %v", err)
			}

			for _, e := range tt.seedEntries {
				if err := client.InsertEntry(ctx, e); err != nil {
					t.Fatalf("failed to seed entry: %v", err)
				}
			}

			fetched, total, err := client.GetEntriesByTallyID(ctx, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetEntriesByTallyID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if total != tt.wantTotal {
				t.Errorf("GetEntriesByTallyID() total = %d, want %d", total, tt.wantTotal)
			}
			if len(fetched) != tt.wantCount {
				t.Errorf("GetEntriesByTallyID() fetched count = %d, want %d", len(fetched), tt.wantCount)
			}
		})
	}
}
