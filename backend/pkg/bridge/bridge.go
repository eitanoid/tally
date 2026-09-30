// Package bridge provides foreign function interface (FFI) bindings for mobile and external callers.
package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/eitanoid/tally/generated/pb/tallyv1"
	"github.com/eitanoid/tally/internal/repository"
	"github.com/eitanoid/tally/internal/schemas"
	"github.com/eitanoid/tally/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// bridge acts as a stateful entry point for external callers.
type bridge struct {
	service *service.TallyService
	ctx     context.Context
}

var (
	instance *bridge
	mu       sync.RWMutex
)

func get() (*bridge, error) {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return nil, fmt.Errorf("bridge engine not initialized: call Init() first")
	}
	return instance, nil
}

// New creates a new bridge stateful client.
func New(dbPath string) error {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return nil // Already initialized
	}

	sqlClient, err := repository.NewSQLiteClient(dbPath)
	if err != nil {
		return fmt.Errorf("failed to init db at %s: %w", dbPath, err)
	}

	instance = &bridge{
		service: service.NewTallyService(sqlClient),
		ctx:     context.Background(),
	}

	return nil
}

// Close gracefully stops the bridge and resets the singleton instance.
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		return nil
	}

	var closeErr error
	if instance.service != nil {
		closeErr = instance.service.Close()
	}

	instance = nil
	return closeErr
}

// Ping is a test function
func Ping(name string) string {
	return fmt.Sprintf("Hello %s! Go engine is alive 🚀", name)
}

// CreateSchema accepts a serialized CreateSchemaRequest and returns a serialized CreateSchemaResponse.
func CreateSchema(requestBytes []byte) []byte {
	return handleRPC(
		requestBytes,
		&tallyv1.CreateSchemaRequest{},
		func(st *bridge, req *tallyv1.CreateSchemaRequest) (*tallyv1.CreateSchemaResponse, error) {
			sr := schemas.NewSchemaRequest(req.Name, req.Description)
			for _, field := range req.GetFields() {
				if field == nil {
					continue
				}
				fieldType, err := mapProtoTypeToDomain(field.GetType())
				if err != nil {
					return nil, fmt.Errorf("invalid type %s for field %s: %w", field.GetName(), field.GetType(), err)
				}
				sr.WithField(field.GetName(), field.GetDescription(), fieldType, field.GetRequired())
			}

			schema, err := st.service.CreateSchema(st.ctx, sr)
			if err != nil {
				return nil, err
			}

			return &tallyv1.CreateSchemaResponse{
				TallyId:       schema.TallyID,
				SchemaVersion: int32(schema.Version),
				Code:          tallyv1.ResponseCode_RESPONSE_CODE_OK,
			}, nil
		},
	)
}

// GetLatestSchema accepts a serialized GetLatestSchemaRequest and returns a serialized GetLatestSchemaResponse.
func GetLatestSchema(requestBytes []byte) []byte {
	return handleRPC(
		requestBytes,
		&tallyv1.GetLatestSchemaRequest{},
		func(st *bridge, req *tallyv1.GetLatestSchemaRequest) (*tallyv1.GetLatestSchemaResponse, error) {
			schema, err := st.service.GetLatestSchema(st.ctx, req.GetTallyId())
			if err != nil {
				return nil, err
			}

			return &tallyv1.GetLatestSchemaResponse{
				Schema: &tallyv1.Schema{
					TallyId:       schema.TallyID,
					SchemaVersion: int32(schema.Version),
					Name:          schema.Name,
					Description:   schema.Description,
					JsonSchema:    schema.JSONSchemaRaw,
					CreatedAt:     timestamppb.New(schema.CreatedAt),
				},
				Code: tallyv1.ResponseCode_RESPONSE_CODE_OK,
			}, nil
		},
	)
}

// ListSchemas accepts a serialized ListSchemasRequest and returns a serialized ListSchemasResponse.
func ListSchemas(requestBytes []byte) []byte {
	return handleRPC(
		requestBytes,
		&tallyv1.ListSchemasRequest{},
		func(st *bridge, req *tallyv1.ListSchemasRequest) (*tallyv1.ListSchemasResponse, error) {
			schemaList, err := st.service.ListSchemas(st.ctx)
			if err != nil {
				return nil, err
			}

			protoSchemas := make([]*tallyv1.Schema, 0, len(schemaList))
			for _, s := range schemaList {
				protoSchemas = append(protoSchemas, &tallyv1.Schema{
					TallyId:       s.TallyID,
					SchemaVersion: int32(s.Version),
					Name:          s.Name,
					Description:   s.Description,
					JsonSchema:    s.JSONSchemaRaw,
					CreatedAt:     timestamppb.New(s.CreatedAt),
				})
			}

			return &tallyv1.ListSchemasResponse{
				Schemas: protoSchemas,
				Code:    tallyv1.ResponseCode_RESPONSE_CODE_OK,
			}, nil
		},
	)
}

// RecordEntry accepts a serialized RecordEntryRequest and returns a serialized RecordEntryResponse.
func RecordEntry(requestBytes []byte) []byte {
	return handleRPC(
		requestBytes,
		&tallyv1.RecordEntryRequest{},
		func(st *bridge, req *tallyv1.RecordEntryRequest) (*tallyv1.RecordEntryResponse, error) {
			entry, err := st.service.RecordEntry(st.ctx, req.GetTallyId(), req.GetPayloadJson())
			if err != nil {
				return nil, err
			}

			return &tallyv1.RecordEntryResponse{
				EntryId: entry.ID,
				Code:    tallyv1.ResponseCode_RESPONSE_CODE_OK,
			}, nil
		},
	)
}

// ListEntries accepts a serialized ListEntriesRequest and returns a serialized ListEntriesResponse.
func ListEntries(requestBytes []byte) []byte {
	return handleRPC(
		requestBytes,
		&tallyv1.ListEntriesRequest{},
		func(st *bridge, req *tallyv1.ListEntriesRequest) (*tallyv1.ListEntriesResponse, error) {
			paginated, err := st.service.ListEntries(st.ctx, req.GetTallyId(), int(req.GetLimit()), int(req.GetOffset()))
			if err != nil {
				return nil, err
			}

			protoEntries := make([]*tallyv1.Entry, 0, len(paginated.Entries))
			for _, e := range paginated.Entries {
				protoEntries = append(protoEntries, &tallyv1.Entry{
					EntryId:       e.ID,
					TallyId:       e.TallyID,
					SchemaVersion: int32(e.SchemaVersion),
					Data:          e.Data,
					CreatedAt:     timestamppb.New(e.CreatedAt),
					UpdatedAt:     timestamppb.New(e.UpdatedAt),
				})
			}

			return &tallyv1.ListEntriesResponse{
				Entries:    protoEntries,
				TotalCount: int32(paginated.TotalCount),
				Limit:      req.Limit,
				Offset:     req.Offset,
				HasMore:    paginated.HasMore,
				Code:       tallyv1.ResponseCode_RESPONSE_CODE_OK,
			}, nil
		},
	)
}

// UpdateEntry accepts a serialized UpdateEntryRequest and returns a serialized UpdateEntryResponse.
func UpdateEntry(requestBytes []byte) []byte {
	return handleRPC(
		requestBytes,
		&tallyv1.UpdateEntryRequest{},
		func(st *bridge, req *tallyv1.UpdateEntryRequest) (*tallyv1.UpdateEntryResponse, error) {
			updatedEntry, err := st.service.UpdateEntry(st.ctx, req.GetEntryId(), req.GetPatchData())
			if err != nil {
				return nil, err
			}

			return &tallyv1.UpdateEntryResponse{
				UpdatedEntry: &tallyv1.Entry{
					EntryId:       updatedEntry.ID,
					TallyId:       updatedEntry.TallyID,
					SchemaVersion: int32(updatedEntry.SchemaVersion),
					Data:          updatedEntry.Data,
					CreatedAt:     timestamppb.New(updatedEntry.CreatedAt),
					UpdatedAt:     timestamppb.New(updatedEntry.UpdatedAt),
				},
				Code: tallyv1.ResponseCode_RESPONSE_CODE_OK,
			}, nil
		},
	)
}

// DeleteEntry accepts a serialized DeleteEntryRequest and returns a serialized DeleteEntryResponse.
func DeleteEntry(requestBytes []byte) []byte {
	return handleRPC(
		requestBytes,
		&tallyv1.DeleteEntryRequest{},
		func(st *bridge, req *tallyv1.DeleteEntryRequest) (*tallyv1.DeleteEntryResponse, error) {
			if err := st.service.DeleteEntry(st.ctx, req.GetEntryId()); err != nil {
				return nil, err
			}

			return &tallyv1.DeleteEntryResponse{
				Code: tallyv1.ResponseCode_RESPONSE_CODE_OK,
			}, nil
		},
	)
}

// DeleteTally accepts a serialized DeleteTallyRequest and returns a serialized DeleteTallyResponse.
func DeleteTally(requestBytes []byte) []byte {
	return handleRPC(
		requestBytes,
		&tallyv1.DeleteTallyRequest{},
		func(st *bridge, req *tallyv1.DeleteTallyRequest) (*tallyv1.DeleteTallyResponse, error) {
			if err := st.service.DeleteTally(st.ctx, req.GetTallyId()); err != nil {
				return nil, err
			}

			return &tallyv1.DeleteTallyResponse{
				Code: tallyv1.ResponseCode_RESPONSE_CODE_OK,
			}, nil
		},
	)
}
