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
	"google.golang.org/protobuf/proto"
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

// ListEntries accepts a serialized ListEntriesRequest and returns a serialized ListEntriesResponse.
func ListEntries(requestBytes []byte) []byte {
	// safely retrieve shared service
	st, err := get()
	if err != nil {
		return marshalProtoError(
			setListEntriesErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	var req tallyv1.ListEntriesRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setListEntriesErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	paginated, err := st.service.ListEntries(st.ctx, req.GetTallyId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return marshalProtoError(
			setListEntriesErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	// Map domain entities to Protobuf response
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

	resp := &tallyv1.ListEntriesResponse{
		Entries:    protoEntries,
		TotalCount: int32(paginated.TotalCount),
		Limit:      req.Limit,
		Offset:     req.Offset,
		HasMore:    paginated.HasMore,
		Code:       tallyv1.ResponseCode_RESPONSE_CODE_OK,
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalProtoError(
			setListEntriesErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			fmt.Errorf("failed to marshal response: %w", err),
		)
	}

	return out
}

// CreateSchema accepts a serialized CreateSchemaRequest and returns a serialized CreateSchemaResponse.
func CreateSchema(requestBytes []byte) []byte {
	// safely retrieve shared service
	st, err := get()
	if err != nil {
		return marshalProtoError(
			setCreateSchemaErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	var req tallyv1.CreateSchemaRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setCreateSchemaErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}
	sr := schemas.NewSchemaRequest(req.Name, req.Description)
	for _, field := range req.GetFields() {
		if field == nil {
			continue
		}
		fieldType, err := mapProtoTypeToDomain(field.GetType())
		if err != nil {
			return marshalProtoError(
				setCreateSchemaErr,
				tallyv1.ResponseCode_RESPONSE_CODE_INVALID_PAYLOAD,
				fmt.Errorf("invalid type %s for field %s", field.GetName(), field.GetType()),
			)

		}
		sr.WithField(field.GetName(), field.GetDescription(), fieldType, field.GetRequired())
	}

	schema, err := st.service.CreateSchema(st.ctx, sr)
	if err != nil {
		return marshalProtoError(
			setCreateSchemaErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	resp := &tallyv1.CreateSchemaResponse{
		TallyId:       schema.TallyID,
		SchemaVersion: int32(schema.Version),
		Code:          tallyv1.ResponseCode_RESPONSE_CODE_OK,
		ErrorMessage:  "",
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalProtoError(
			setCreateSchemaErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			fmt.Errorf("failed to marshal response: %w", err),
		)
	}
	return out
}

// GetLatestSchema accepts a serialized GetLatestSchemaRequest and returns a serialized GetLatestSchemaResponse.
func GetLatestSchema(requestBytes []byte) []byte {
	// safely retrieve shared service
	st, err := get()
	if err != nil {
		return marshalProtoError(
			setGetLatestSchemaErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	var req tallyv1.GetLatestSchemaRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setGetLatestSchemaErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	schema, err := st.service.GetLatestSchema(st.ctx, req.GetTallyId())
	if err != nil {
		return marshalProtoError(
			setGetLatestSchemaErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	resp := &tallyv1.GetLatestSchemaResponse{
		Schema: &tallyv1.Schema{
			TallyId:       schema.TallyID,
			SchemaVersion: int32(schema.Version),
			Name:          schema.Name,
			Description:   schema.Description,
			JsonSchema:    schema.JSONSchemaRaw,
			CreatedAt:     timestamppb.New(schema.CreatedAt),
		},
		Code: tallyv1.ResponseCode_RESPONSE_CODE_OK,
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalProtoError(
			setGetLatestSchemaErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			fmt.Errorf("failed to marshal response: %w", err),
		)
	}
	return out
}

// ListSchemas accepts a serialized ListSchemasRequest and returns a serialized ListSchemasResponse.
func ListSchemas(requestBytes []byte) []byte {
	// safely retrieve shared service
	st, err := get()
	if err != nil {
		return marshalProtoError(
			setListSchemasErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	var req tallyv1.ListSchemasRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setListSchemasErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	schemaList, err := st.service.ListSchemas(st.ctx)
	if err != nil {
		return marshalProtoError(
			setListSchemasErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
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

	resp := &tallyv1.ListSchemasResponse{
		Schemas: protoSchemas,
		Code:    tallyv1.ResponseCode_RESPONSE_CODE_OK,
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalProtoError(
			setListSchemasErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			fmt.Errorf("failed to marshal response: %w", err),
		)
	}
	return out
}

// RecordEntry accepts a serialized RecordEntryRequest and returns a serialized RecordEntryResponse.
func RecordEntry(requestBytes []byte) []byte {
	// safely retrieve shared service
	st, err := get()
	if err != nil {
		return marshalProtoError(
			setListSchemasErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	var req tallyv1.RecordEntryRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setRecordEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	entry, err := st.service.RecordEntry(st.ctx, req.GetTallyId(), req.GetPayloadJson())
	if err != nil {
		return marshalProtoError(
			setRecordEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	resp := &tallyv1.RecordEntryResponse{
		EntryId: entry.ID,
		Code:    tallyv1.ResponseCode_RESPONSE_CODE_OK,
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalProtoError(
			setRecordEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			fmt.Errorf("failed to marshal response: %w", err),
		)
	}
	return out
}

// UpdateEntry accepts a serialized UpdateEntryRequest and returns a serialized UpdateEntryResponse.
func UpdateEntry(requestBytes []byte) []byte {
	st, err := get()
	if err != nil {
		return marshalProtoError(
			setUpdateEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	var req tallyv1.UpdateEntryRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setUpdateEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	updatedEntry, err := st.service.UpdateEntry(st.ctx, req.GetEntryId(), req.GetPatchData())
	if err != nil {
		return marshalProtoError(
			setUpdateEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	resp := &tallyv1.UpdateEntryResponse{
		// Entry: updatedEntry,
		UpdatedEntry: &tallyv1.Entry{
			EntryId:       updatedEntry.ID,
			TallyId:       updatedEntry.TallyID,
			SchemaVersion: int32(updatedEntry.SchemaVersion),
			Data:          updatedEntry.Data,
			CreatedAt:     timestamppb.New(updatedEntry.CreatedAt),
			UpdatedAt:     timestamppb.New(updatedEntry.UpdatedAt),
		},
		Code: tallyv1.ResponseCode_RESPONSE_CODE_OK,
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalProtoError(
			setUpdateEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			fmt.Errorf("failed to marshal response: %w", err),
		)
	}
	return out
}

// DeleteEntry accepts a serialized DeleteEntryRequest and returns a serialized DeleteEntryResponse.
func DeleteEntry(requestBytes []byte) []byte {
	// safely retrieve shared service
	st, err := get()
	if err != nil {
		return marshalProtoError(
			setDeleteEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}
	var req tallyv1.DeleteEntryRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setDeleteEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	if err := st.service.DeleteEntry(st.ctx, req.GetEntryId()); err != nil {
		return marshalProtoError(
			setDeleteEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	resp := &tallyv1.DeleteEntryResponse{
		Code: tallyv1.ResponseCode_RESPONSE_CODE_OK,
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalProtoError(
			setDeleteEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			fmt.Errorf("failed to marshal response: %w", err),
		)
	}
	return out
}

// DeleteTally accepts a serialized DeleteTallyRequest and returns a serialized DeleteTallyResponse.
func DeleteTally(requestBytes []byte) []byte {
	// safely retrieve shared service
	st, err := get()
	if err != nil {
		return marshalProtoError(
			setDeleteTallyErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}
	var req tallyv1.DeleteTallyRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setDeleteTallyErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	if err := st.service.DeleteTally(st.ctx, req.GetTallyId()); err != nil {
		return marshalProtoError(
			setDeleteTallyErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	resp := &tallyv1.DeleteTallyResponse{
		Code: tallyv1.ResponseCode_RESPONSE_CODE_OK,
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalProtoError(
			setDeleteTallyErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			fmt.Errorf("failed to marshal response: %w", err),
		)
	}
	return out
}
