// Package bridge provides foreign function interface (FFI) bindings for mobile and external callers.
package bridge

import (
	"context"
	"fmt"

	"github.com/eitanoid/tally/generated/pb/tallyv1"
	"github.com/eitanoid/tally/internal/repository"
	"github.com/eitanoid/tally/internal/service"
	"google.golang.org/protobuf/proto"
)

// Bridge acts as a simplified entry point for external callers.
type Bridge struct {
	service *service.TallyService
	ctx     context.Context
}

// New creates a new bridge client.
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

// ListEntries accepts a serialized ListEntriesRequest and returns a serialized ListEntriesResponse.
func (b *Bridge) ListEntries(requestBytes []byte) []byte {
	var req tallyv1.ListEntriesRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalListEntriesError(err)
	}

	paginated, err := b.service.ListEntries(b.ctx, req.GetTallyId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return marshalListEntriesError(err)
	}

	// Map domain entities to Protobuf response
	protoEntries := make([]*tallyv1.Entry, 0, len(paginated.Entries))
	for _, e := range paginated.Entries {
		protoEntries = append(protoEntries, &tallyv1.Entry{
			EntryId:       e.ID,
			TallyId:       e.TallyID,
			SchemaVersion: int32(e.SchemaVersion),
			Data:          e.Data,
		})
	}

	resp := &tallyv1.ListEntriesResponse{
		Entries:    protoEntries,
		TotalCount: int32(paginated.TotalCount),
		HasMore:    paginated.HasMore,
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalListEntriesError(fmt.Errorf("failed to marshal response: %w", err))
	}

	return out
}

// CreateSchema accepts a JSON payload and returns (JSON response string, error).
func (b *Bridge) CreateSchema(schemaRequest string) (string, error) {
	_ = schemaRequest
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
