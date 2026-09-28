// Package bridge provides foreign function interface (FFI) bindings for mobile and external callers.
package bridge

import (
	"context"
	"fmt"

	"github.com/eitanoid/tally/generated/pb/tallyv1"
	"github.com/eitanoid/tally/internal/repository"
	"github.com/eitanoid/tally/internal/schemas"
	"github.com/eitanoid/tally/internal/service"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		return marshalProtoError(
			setListEntriesErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	paginated, err := b.service.ListEntries(b.ctx, req.GetTallyId(), int(req.GetLimit()), int(req.GetOffset()))
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
func (b *Bridge) CreateSchema(requestBytes []byte) []byte {
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

	schema, err := b.service.CreateSchema(b.ctx, sr)
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
func (b *Bridge) GetLatestSchema(requestBytes []byte) []byte {
	var req tallyv1.GetLatestSchemaRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setGetLatestSchemaErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	schema, err := b.service.GetLatestSchema(b.ctx, req.GetTallyId())
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
func (b *Bridge) ListSchemas(requestBytes []byte) []byte {
	var req tallyv1.ListSchemasRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setListSchemasErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	schemaList, err := b.service.ListSchemas(b.ctx)
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
func (b *Bridge) RecordEntry(requestBytes []byte) []byte {
	var req tallyv1.RecordEntryRequest
	if err := unmarshalRequest(requestBytes, &req); err != nil {
		return marshalProtoError(
			setRecordEntryErr,
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			err,
		)
	}

	entry, err := b.service.RecordEntry(b.ctx, req.GetTallyId(), req.GetPayloadJson())
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
