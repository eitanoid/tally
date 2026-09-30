package bridge

import (
	"fmt"
	"reflect"

	"github.com/eitanoid/tally/generated/pb/tallyv1"
	"github.com/eitanoid/tally/internal/schemas"
	"google.golang.org/protobuf/proto"
)

// mapProtoTypeToDomain maps proto FieldType enums to domain SupportedType.
func mapProtoTypeToDomain(pt tallyv1.FieldFormat) (schemas.SupportedType, error) {
	switch pt {
	case tallyv1.FieldFormat_FIELD_FORMAT_STRING:
		return schemas.TypeString, nil
	case tallyv1.FieldFormat_FIELD_FORMAT_BOOLEAN:
		return schemas.TypeBool, nil
	case tallyv1.FieldFormat_FIELD_FORMAT_INTEGER:
		return schemas.TypeInt, nil
	case tallyv1.FieldFormat_FIELD_FORMAT_NUMBER:
		return schemas.TypeFloat, nil
	case tallyv1.FieldFormat_FIELD_FORMAT_DATE_TIME:
		return schemas.TypeDateTime, nil
	case tallyv1.FieldFormat_FIELD_FORMAT_DATE:
		return schemas.TypeDate, nil
	case tallyv1.FieldFormat_FIELD_FORMAT_TIME:
		return schemas.TypeTime, nil
	case tallyv1.FieldFormat_FIELD_FORMAT_DURATION:
		return schemas.TypeDuration, nil
	default:
		return "", fmt.Errorf("unsupported or unspecified proto field type: %v", pt)
	}
}

func handleRPC[Req proto.Message, Resp proto.Message](
	requestBytes []byte,
	req Req,
	action func(st *bridge, req Req) (Resp, error),
) []byte {
	st, err := get()
	if err != nil {
		return marshalProtoError[Resp](tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR, err)
	}

	if err := unmarshalRequest(requestBytes, req); err != nil {
		return marshalProtoError[Resp](tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR, err)
	}

	resp, err := action(st, req)
	if err != nil {
		return marshalProtoError[Resp](tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR, err)
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return marshalProtoError[Resp](
			tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			fmt.Errorf("failed to marshal response: %w", err),
		)
	}

	return out
}

// unmarshalRequest unmarshals incoming binary bytes into a proto message.
func unmarshalRequest[T proto.Message](data []byte, msg T) error {
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}
	return nil
}

// marshalProtoError instantiates a response value, applies error details, and returns binary bytes.
func marshalProtoError[Resp proto.Message](code tallyv1.ResponseCode, err error) []byte {
	// Dynamically instantiate a new pointer to Resp struct (e.g. *tallyv1.DeleteEntryResponse)
	respType := reflect.TypeOf((*Resp)(nil)).Elem()
	resp := reflect.New(respType.Elem()).Interface().(Resp)

	setErrorDetails(resp, code, err)

	out, marshalErr := proto.Marshal(resp)
	if marshalErr != nil {
		return []byte{}
	}
	return out
}

func setErrorDetails[T proto.Message](resp T, code tallyv1.ResponseCode, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	switch r := any(resp).(type) {
	case *tallyv1.CreateSchemaResponse:
		r.Code, r.ErrorMessage = code, errMsg
	case *tallyv1.ListSchemasResponse:
		r.Code, r.ErrorMessage = code, errMsg
	case *tallyv1.GetLatestSchemaResponse:
		r.Code, r.ErrorMessage = code, errMsg
	case *tallyv1.ListEntriesResponse:
		r.Code, r.ErrorMessage = code, errMsg
	case *tallyv1.RecordEntryResponse:
		r.Code, r.ErrorMessage = code, errMsg
	case *tallyv1.DeleteEntryResponse:
		r.Code, r.ErrorMessage = code, errMsg
	case *tallyv1.DeleteTallyResponse:
		r.Code, r.ErrorMessage = code, errMsg
	case *tallyv1.UpdateEntryResponse:
		r.Code, r.ErrorMessage = code, errMsg
	}
}
