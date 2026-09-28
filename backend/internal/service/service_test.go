// Package service_test verifies business logic orchestration in TallyService.
package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/eitanoid/tally/internal/entries"
	"github.com/eitanoid/tally/internal/repository"
	"github.com/eitanoid/tally/internal/schemas"
	"github.com/eitanoid/tally/internal/service"
)

func setupTestService(t *testing.T) (*service.TallyService, repository.Repository) {
	t.Helper()

	client, err := repository.NewSQLiteClient(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	return service.NewTallyService(client), client
}

func TestCreateSchema(t *testing.T) {
	tests := []struct {
		name    string
		buildFn func() *schemas.SchemaRequest
		wantErr bool
	}{
		{
			name: "valid schema request creation",
			buildFn: func() *schemas.SchemaRequest {
				req := schemas.NewSchemaRequest("Water Intake", "Track daily hydration")
				_ = req.WithField("ml", "", "number", true)
				return req
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := setupTestService(t)
			ctx := t.Context()

			req := tt.buildFn()
			got, err := svc.CreateSchema(ctx, req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateSchema() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if got.TallyID == "" {
					t.Errorf("expected TallyID to be auto-generated, got empty string")
				}
				if got.Version < 1 {
					t.Errorf("expected Version to be >= 1, got %d", got.Version)
				}
				if got.CreatedAt.IsZero() {
					t.Errorf("expected CreatedAt timestamp to be populated after persistence")
				}
			}
		})
	}
}

func TestGetLatestSchema(t *testing.T) {
	tests := []struct {
		name        string
		seedFn      func(ctx context.Context, svc *service.TallyService) (string, int)
		queryID     string
		wantVersion int
		wantErrIs   error
	}{
		{
			name: "fetches active generated schema",
			seedFn: func(ctx context.Context, svc *service.TallyService) (string, int) {
				req := schemas.NewSchemaRequest("Habits Tracker", "Track daily habits")
				_ = req.WithField("completed", "", "boolean", true)

				s, err := svc.CreateSchema(ctx, req)
				if err != nil {
					t.Fatalf("failed to seed schema: %v", err)
				}
				return s.TallyID, s.Version
			},
			wantErrIs: nil,
		},
		{
			name: "non-existent tally_id error",
			seedFn: func(_ context.Context, _ *service.TallyService) (string, int) {
				return "00000000-0000-0000-0000-000000000000", 0
			},
			wantErrIs: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := setupTestService(t)
			ctx := t.Context()

			tallyID, expectedVersion := tt.seedFn(ctx, svc)

			got, err := svc.GetLatestSchema(ctx, tallyID)
			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("GetLatestSchema() error = %v, wantErrIs %v", err, tt.wantErrIs)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Version != expectedVersion {
				t.Errorf("GetLatestSchema().Version = %d, want %d", got.Version, expectedVersion)
			}
		})
	}
}

func TestRecordEntry(t *testing.T) {
	tests := []struct {
		name     string
		seedFn   func(ctx context.Context, svc *service.TallyService) string
		targetID string
		rawData  string
		wantErr  bool
	}{
		{
			name: "successfully record valid entry against generated schema",
			seedFn: func(ctx context.Context, svc *service.TallyService) string {
				req := schemas.NewSchemaRequest("Workout Tracker", "Gym log")
				_ = req.WithField("reps", "", "integer", true)
				s, err := svc.CreateSchema(ctx, req)
				if err != nil {
					t.Fatalf("failed to seed schema: %v", err)
				}
				return s.TallyID
			},
			rawData: `{"reps": 12}`,
			wantErr: false,
		},
		{
			name: "validation failure against generated schema",
			seedFn: func(ctx context.Context, svc *service.TallyService) string {
				req := schemas.NewSchemaRequest("Workout Tracker", "Gym log")
				_ = req.WithField("reps", "", "integer", true)
				s, err := svc.CreateSchema(ctx, req)
				if err != nil {
					t.Fatalf("failed to seed schema: %v", err)
				}
				return s.TallyID
			},
			rawData: `{"reps": "twelve"}`,
			wantErr: true,
		},
		{
			name: "failed when tally schema does not exist",
			seedFn: func(_ context.Context, _ *service.TallyService) string {
				return "missing-tally-id"
			},
			rawData: `{"reps": 12}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := setupTestService(t)
			ctx := t.Context()

			tallyID := tt.seedFn(ctx, svc)

			got, err := svc.RecordEntry(ctx, tallyID, tt.rawData)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RecordEntry() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if got.ID == "" {
					t.Errorf("expected KSUID ID to be set on recorded entry")
				}
				if got.CreatedAt.IsZero() {
					t.Errorf("expected CreatedAt timestamp to be populated after persistence")
				}
			}
		})
	}
}

func TestListEntries(t *testing.T) {
	tests := []struct {
		name        string
		seedEntries []string
		limit       int
		offset      int
		wantCount   int
		wantTotal   int
		wantHasMore bool
		wantErr     bool
	}{
		{
			name:        "paginates entries with has_more true",
			seedEntries: []string{`{"ml": 250}`, `{"ml": 500}`, `{"ml": 750}`},
			limit:       2,
			offset:      0,
			wantCount:   2,
			wantTotal:   3,
			wantHasMore: true,
			wantErr:     false,
		},
		{
			name:        "paginates last page with has_more false",
			seedEntries: []string{`{"ml": 250}`, `{"ml": 500}`, `{"ml": 750}`},
			limit:       2,
			offset:      2,
			wantCount:   1,
			wantTotal:   3,
			wantHasMore: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := setupTestService(t)
			ctx := context.Background()

			// Create schema dynamically using builder
			req := schemas.NewSchemaRequest("Hydration", "Daily fluid intake")
			_ = req.WithField("ml", "", "number", true)

			s, err := svc.CreateSchema(ctx, req)
			if err != nil {
				t.Fatalf("failed to create schema: %v", err)
			}

			for _, raw := range tt.seedEntries {
				if _, err := svc.RecordEntry(ctx, s.TallyID, raw); err != nil {
					t.Fatalf("failed to record entry: %v", err)
				}
			}

			res, err := svc.ListEntries(ctx, s.TallyID, tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListEntries() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if res.TotalCount != tt.wantTotal {
					t.Errorf("ListEntries().TotalCount = %d, want %d", res.TotalCount, tt.wantTotal)
				}
				if len(res.Entries) != tt.wantCount {
					t.Errorf("ListEntries().Entries count = %d, want %d", len(res.Entries), tt.wantCount)
				}
				if res.HasMore != tt.wantHasMore {
					t.Errorf("ListEntries().HasMore = %v, want %v", res.HasMore, tt.wantHasMore)
				}
			}
		})
	}
}

func TestTallyService_UpdateEntry(t *testing.T) {
	tallyUUID := "550e8400-e29b-41d4-a716-446655440000"
	validSchemaJSON := `{"type":"object","properties":{"count":{"type":"integer"}}}`

	tests := []struct {
		name       string
		entryID    string
		patchData  string
		seedSchema *schemas.TallySchema
		seedEntry  *entries.TallyEntry
		wantErrIs  error
		wantData   string
	}{
		{
			name:      "successful merge patch update",
			entryID:   "entry-1",
			patchData: `{"count": 10}`,
			seedSchema: &schemas.TallySchema{
				TallyID:       tallyUUID,
				Version:       1,
				Name:          "Counter",
				JSONSchemaRaw: validSchemaJSON,
			},
			seedEntry: &entries.TallyEntry{
				ID:            "entry-1",
				TallyID:       tallyUUID,
				SchemaVersion: 1,
				Data:          `{"count": 5}`,
			},
			wantErrIs: nil,
			wantData:  `{"count":10}`,
		},
		{
			name:       "returns ErrNotFound when entry does not exist",
			entryID:    "non-existent-entry",
			patchData:  `{"count": 10}`,
			seedSchema: nil,
			seedEntry:  nil,
			wantErrIs:  entries.ErrNotFound,
		},
		{
			name:      "returns ErrInvalidPatch on malformed JSON payload",
			entryID:   "entry-1",
			patchData: `invalid-json-payload`,
			seedSchema: &schemas.TallySchema{
				TallyID:       tallyUUID,
				Version:       1,
				Name:          "Counter",
				JSONSchemaRaw: validSchemaJSON,
			},
			seedEntry: &entries.TallyEntry{
				ID:            "entry-1",
				TallyID:       tallyUUID,
				SchemaVersion: 1,
				Data:          `{"count": 5}`,
			},
			wantErrIs: entries.ErrInvalidPatch,
		},
		{
			name:      "returns ErrInvalidData when merge patch violates JSON schema",
			entryID:   "entry-1",
			patchData: `{"count": "not-an-integer"}`,
			seedSchema: &schemas.TallySchema{
				TallyID:       tallyUUID,
				Version:       1,
				Name:          "Counter",
				JSONSchemaRaw: validSchemaJSON,
			},
			seedEntry: &entries.TallyEntry{
				ID:            "entry-1",
				TallyID:       tallyUUID,
				SchemaVersion: 1,
				Data:          `{"count": 5}`,
			},
			wantErrIs: entries.ErrInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := setupTestService(t)
			ctx := context.Background()

			if tt.seedSchema != nil {
				if err := repo.InsertSchema(ctx, tt.seedSchema); err != nil {
					t.Fatalf("failed to seed schema: %v", err)
				}
			}

			if tt.seedEntry != nil {
				if err := repo.InsertEntry(ctx, tt.seedEntry); err != nil {
					t.Fatalf("failed to seed entry: %v", err)
				}
			}

			updated, err := svc.UpdateEntry(ctx, tt.entryID, tt.patchData)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("UpdateEntry() error = %v, wantErrIs %v", err, tt.wantErrIs)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if updated.Data != tt.wantData {
				t.Errorf("UpdateEntry() returned Data = %s, want %s", updated.Data, tt.wantData)
			}

			persisted, err := repo.GetEntryByID(ctx, tt.entryID)
			if err != nil {
				t.Fatalf("failed to fetch persisted entry: %v", err)
			}
			if persisted.Data != tt.wantData {
				t.Errorf("DB persisted Data = %s, want %s", persisted.Data, tt.wantData)
			}
		})
	}
}
