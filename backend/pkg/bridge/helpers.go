package bridge

import (
	"fmt"

	"github.com/eitanoid/tally/generated/pb/tallyv1"
	"google.golang.org/protobuf/proto"
)

// unmarshalRequest unmarshals incoming binary bytes into a proto message.
func unmarshalRequest[T proto.Message](data []byte, msg T) error {
	if len(data) == 0 {
		return fmt.Errorf("received empty payload")
	}
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}
	return nil
}

// marshalError creates a serialized ListEntriesResponse containing an error string.
func marshalListEntriesError(err error) []byte {
	res := &tallyv1.ListEntriesResponse{
		ErrorMessage: err.Error(),
	}
	out, _ := proto.Marshal(res)
	return out
}

// marshalSchemaError creates a serialized CreateSchemaResponse containing an error string.
func marshalCreateSchemaError(err error) []byte {
	res := &tallyv1.CreateSchemaResponse{
		ErrorMessage: err.Error(),
	}
	out, _ := proto.Marshal(res)
	return out
}
