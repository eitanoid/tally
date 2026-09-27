package bridge

import (
	"context"
	"fmt"

	"github.com/eitanoid/tally/internal/repository"
	"github.com/eitanoid/tally/internal/service"
)

type Bridge struct {
	service *service.TallyService
	ctx     context.Context
}

func New(dbPath string) (*Bridge, error) {
	sqlClient, err := repository.NewSQLiteClient(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init db at %s: %w", dbPath, err)
	}

	return &Bridge{
		service: service.NewTallyService(sqlClient),
		ctx:     context.Background(),
	}, nil
}

// ListEntries returns (JSON response string, error).
func (b *Bridge) ListEntries(tallyId string, limit, offset int32) (string, error) {
	if b.service == nil {
		return "", fmt.Errorf("service not initialized")
	}

	// return jsonString, nil
	return "{}", nil
}

// CreateSchema accepts a JSON payload and returns (JSON response string, error).
func (b *Bridge) CreateSchema(schemaRequest string) (string, error) {
	if b.service == nil {
		return "", fmt.Errorf("service not initialized")
	}

	return "{}", nil
}

// GetLatestSchema accepts a tallyID and returns (JSON response string, error).
func (b *Bridge) GetLatestSchema(tallyID string) (string, error) {
	_, _ = b.service.GetLatestSchema(b.ctx, tallyID)
	return "", nil
}

// ListSchemas returns all tallys and their latest versions. (JSON, error)
func (b *Bridge) ListSchemas() (string, error) {
	_, _ = b.service.ListSchemas(b.ctx)
	return "", nil
}

// RecordEntry fetches the active schema for tallyID, validates payload against it,
// constructs the domain entry, and saves it to SQLite.
// returns JSON, error
func (b *Bridge) RecordEntry(tallyID string, rawData string) (string, error) {
	_, _ = b.service.RecordEntry(b.ctx, tallyID, rawData)
	return "", nil
}
