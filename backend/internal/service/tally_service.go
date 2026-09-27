package service

import (
	"context"
	"fmt"

	"github.com/eitanoid/tally/internal/entries"
	"github.com/eitanoid/tally/internal/repository"
	"github.com/eitanoid/tally/internal/schemas"
)

type TallyService struct {
	repo repository.Repository
}

func NewTallyService(repo repository.Repository) *TallyService {
	return &TallyService{
		repo: repo,
	}
}

type PaginatedResult struct {
	Entries    []entries.TallyEntry `json:"entries"`
	TotalCount int                  `json:"total_count"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	HasMore    bool                 `json:"has_more"`
}

// CreateSchema builds a new tally schema and persists it.
func (s *TallyService) CreateSchema(ctx context.Context, req *schemas.SchemaRequest) (*schemas.TallySchema, error) {

	schema, err := req.Build()
	if err != nil {
		return nil, fmt.Errorf("service failed to create schema domain object: %w", err)
	}

	if err := s.repo.InsertSchema(ctx, schema); err != nil {
		return nil, fmt.Errorf("service failed to persist schema: %w", err)
	}

	return schema, nil
}

// GetLatestSchema retrieves the active schema definition for a tally ID.
func (s *TallyService) GetLatestSchema(ctx context.Context, tallyID string) (*schemas.TallySchema, error) {
	return s.repo.GetLatestSchemaByID(ctx, tallyID)
}

// ListSchemas returns all tallys and their latest versions.
func (s *TallyService) ListSchemas(ctx context.Context) ([]schemas.TallySchema, error) {
	return s.repo.GetAllLatestSchemas(ctx)
}

// RecordEntry fetches the active schema for tallyID, validates payload against it,
// constructs the domain entry, and saves it to SQLite.
func (s *TallyService) RecordEntry(ctx context.Context, tallyID string, rawData string) (*entries.TallyEntry, error) {
	// Fetch latest schema version for this tally from DB
	activeSchema, err := s.repo.GetLatestSchemaByID(ctx, tallyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active schema for tally '%s': %w", tallyID, err)
	}

	// Create and validate entry
	entry, err := entries.CreateTallyEntry(*activeSchema, rawData)
	if err != nil {
		return nil, fmt.Errorf("entry domain creation failed: %w", err)
	}

	if err := s.repo.InsertEntry(ctx, &entry); err != nil {
		return nil, fmt.Errorf("failed to insert entry: %w", err)
	}

	return &entry, nil
}

// ListEntries fetches recorded logs for a given tally up to limit.
func (s *TallyService) ListEntries(ctx context.Context, tallyID string, limit int, offset int) (*PaginatedResult, error) {
	entries, total, err := s.repo.GetEntriesByTallyID(ctx, repository.EntryFilter{
		TallyID: tallyID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}

	hasMore := (offset + len(entries)) < total

	return &PaginatedResult{
		Entries:    entries,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}, nil
}
