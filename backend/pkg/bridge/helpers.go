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

// unmarshalRequest unmarshals incoming binary bytes into a proto message.
func unmarshalRequest[T proto.Message](data []byte, msg T) error {
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}
	return nil
}

// marshalProtoError populates a proto message using a setter callback and returns marshaled bytes.
func marshalProtoError[T proto.Message](setErr func(T, tallyv1.ResponseCode, string), code tallyv1.ResponseCode, err error) []byte {
	var msg T
	// Safely allocate non-nil struct pointer if T is a pointer type (*tallyv1.SomeResponse)
	msgType := reflect.TypeOf(msg)
	if msgType.Kind() == reflect.Pointer {
		msg = reflect.New(msgType.Elem()).Interface().(T)
	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	setErr(msg, code, errMsg)

	out, marshalErr := proto.Marshal(msg)
	if marshalErr != nil {
		return []byte{}
	}
	return out
}

func setCreateSchemaErr(r *tallyv1.CreateSchemaResponse, code tallyv1.ResponseCode, msg string) {
	r.Code = code
	r.ErrorMessage = msg
}
func setListSchemasErr(r *tallyv1.ListSchemasResponse, code tallyv1.ResponseCode, msg string) {
	r.Code = code
	r.ErrorMessage = msg
}
func setGetLatestSchemaErr(r *tallyv1.GetLatestSchemaResponse, code tallyv1.ResponseCode, msg string) {
	r.Code = code
	r.ErrorMessage = msg
}

func setListEntriesErr(r *tallyv1.ListEntriesResponse, code tallyv1.ResponseCode, msg string) {
	r.Code = code
	r.ErrorMessage = msg
}
func setRecordEntryErr(r *tallyv1.RecordEntryResponse, code tallyv1.ResponseCode, msg string) {
	r.Code = code
	r.ErrorMessage = msg
}
