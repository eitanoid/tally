package service

import (
	"context"
	"fmt"

	"github.com/eitanoid/habit-tracker/internal/entries"
	"github.com/eitanoid/habit-tracker/internal/repository"
	"github.com/eitanoid/habit-tracker/internal/schemas"
)

type HabitService struct {
	repo repository.Repository
}

func NewHabitService(repo repository.Repository) *HabitService {
	return &HabitService{
		repo: repo,
	}
}

// CreateSchema builds a new habit schema and persists it.
func (s *HabitService) CreateSchema(ctx context.Context, req *schemas.SchemaRequest) (*schemas.HabitSchema, error) {

	schema, err := req.Create()
	if err != nil {
		return nil, fmt.Errorf("service failed to create schema domain object: %w", err)
	}

	if err := s.repo.InsertSchema(ctx, schema); err != nil {
		return nil, fmt.Errorf("service failed to persist schema: %w", err)
	}

	return schema, nil
}

// GetLatestSchema retrieves the active schema definition for a habit ID.
func (s *HabitService) GetLatestSchema(ctx context.Context, habitID string) (*schemas.HabitSchema, error) {
	return s.repo.GetLatestSchemaByID(ctx, habitID)
}

// ListSchemas returns all habits and their latest versions.
func (s *HabitService) ListSchemas(ctx context.Context) ([]schemas.HabitSchema, error) {
	return s.repo.GetAllLatestSchemas(ctx)
}

// RecordEntry fetches the active schema for habitID, validates payload against it,
// constructs the domain entry, and saves it to SQLite.
func (s *HabitService) RecordEntry(ctx context.Context, habitID string, rawData string) (*entries.HabitEntry, error) {
	// Fetch latest schema version for this habit from DB
	activeSchema, err := s.repo.GetLatestSchemaByID(ctx, habitID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active schema for habit '%s': %w", habitID, err)
	}

	// Create and validate entry
	entry, err := entries.CreateHabitEntry(*activeSchema, rawData)
	if err != nil {
		return nil, fmt.Errorf("entry domain creation failed: %w", err)
	}

	if err := s.repo.InsertEntry(ctx, &entry); err != nil {
		return nil, fmt.Errorf("failed to insert entry: %w", err)
	}

	return &entry, nil
}

// ListEntries fetches recorded logs for a given habit up to limit.
func (s *HabitService) ListEntries(ctx context.Context, habitID string, limit int) ([]entries.HabitEntry, error) {
	return s.repo.GetEntriesByHabitID(ctx, habitID, limit)
}
