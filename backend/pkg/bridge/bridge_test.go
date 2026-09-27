// Package bridge_test tests the foreign function interface (FFI) serialization and business logic boundary.
package bridge_test

import (
	"testing"

	"github.com/eitanoid/tally/generated/pb/tallyv1"
	"github.com/eitanoid/tally/pkg/bridge"
	"google.golang.org/protobuf/proto"
)

func setupTestBridge(t *testing.T) *bridge.Bridge {
	t.Helper()

	// Initialize Bridge with in-memory SQLite database
	b, err := bridge.New(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test bridge: %v", err)
	}

	return b
}

func TestBridge_CreateSchema(t *testing.T) {
	tests := []struct {
		name         string
		req          *tallyv1.CreateSchemaRequest
		corruptBytes []byte
		wantCode     tallyv1.ResponseCode
	}{
		{
			name: "successfully create schema via protobuf payload",
			req: &tallyv1.CreateSchemaRequest{
				Name:        "Books Read",
				Description: "Track reading progress",
				Fields: []*tallyv1.SchemaRequestField{
					{
						Name:        "title",
						Description: "Book title",
						Type:        tallyv1.FieldFormat_FIELD_FORMAT_STRING,
						Required:    true,
					},
					{
						Name:        "pages",
						Description: "Total pages",
						Type:        tallyv1.FieldFormat_FIELD_FORMAT_INTEGER,
						Required:    false,
					},
				},
			},
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_OK,
		},
		{
			name: "reject invalid or unspecified field format",
			req: &tallyv1.CreateSchemaRequest{
				Name:        "Invalid Schema",
				Description: "Invalid field type",
				Fields: []*tallyv1.SchemaRequestField{
					{
						Name: "unknown_field",
						Type: tallyv1.FieldFormat_FIELD_FORMAT_UNSPECIFIED,
					},
				},
			},
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_INVALID_PAYLOAD,
		},
		{
			name:         "handle malformed input bytes gracefully",
			corruptBytes: []byte("invalid-proto-bytes"),
			wantCode:     tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := setupTestBridge(t)

			var inputBytes []byte
			var err error

			if tt.corruptBytes != nil {
				inputBytes = tt.corruptBytes
			} else {
				inputBytes, err = proto.Marshal(tt.req)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
			}

			outBytes := b.CreateSchema(inputBytes)

			var resp tallyv1.CreateSchemaResponse
			if err := proto.Unmarshal(outBytes, &resp); err != nil {
				t.Fatalf("failed to unmarshal response bytes: %v", err)
			}

			if resp.GetCode() != tt.wantCode {
				t.Errorf("CreateSchema() code = %v, want %v (error_message: %s)", resp.GetCode(), tt.wantCode, resp.GetErrorMessage())
			}

			if tt.wantCode == tallyv1.ResponseCode_RESPONSE_CODE_OK {
				if resp.GetTallyId() == "" {
					t.Errorf("expected TallyId to be set, got empty string")
				}
				if resp.GetSchemaVersion() < 1 {
					t.Errorf("expected SchemaVersion >= 1, got %d", resp.GetSchemaVersion())
				}
			}
		})
	}
}

func TestBridge_GetLatestSchema(t *testing.T) {
	b := setupTestBridge(t)

	// Seed schema
	createReq := &tallyv1.CreateSchemaRequest{
		Name:        "Coffee Log",
		Description: "Daily roasts",
		Fields: []*tallyv1.SchemaRequestField{
			{Name: "roast", Type: tallyv1.FieldFormat_FIELD_FORMAT_STRING, Required: true},
		},
	}
	createBytes, _ := proto.Marshal(createReq)
	createOut := b.CreateSchema(createBytes)
	var createResp tallyv1.CreateSchemaResponse
	_ = proto.Unmarshal(createOut, &createResp)

	seededTallyID := createResp.GetTallyId()

	tests := []struct {
		name     string
		tallyID  string
		wantCode tallyv1.ResponseCode
	}{
		{
			name:     "fetch active schema successfully",
			tallyID:  seededTallyID,
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_OK,
		},
		{
			name:     "return error for non-existent tally_id",
			tallyID:  "00000000-0000-0000-0000-000000000000",
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBytes, _ := proto.Marshal(&tallyv1.GetLatestSchemaRequest{TallyId: tt.tallyID})
			outBytes := b.GetLatestSchema(reqBytes)

			var resp tallyv1.GetLatestSchemaResponse
			if err := proto.Unmarshal(outBytes, &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if resp.GetCode() != tt.wantCode {
				t.Errorf("GetLatestSchema() code = %v, want %v", resp.GetCode(), tt.wantCode)
			}

			if tt.wantCode == tallyv1.ResponseCode_RESPONSE_CODE_OK {
				if resp.GetSchema().GetTallyId() != seededTallyID {
					t.Errorf("got TallyId %s, want %s", resp.GetSchema().GetTallyId(), seededTallyID)
				}
				if resp.GetSchema().GetCreatedAt() == nil {
					t.Errorf("expected CreatedAt timestamp to be populated in proto schema")
				}
			}
		})
	}
}

func TestBridge_ListSchemas(t *testing.T) {
	b := setupTestBridge(t)

	// Seed 2 schemas
	for _, name := range []string{"Habits", "Finances"} {
		reqBytes, _ := proto.Marshal(&tallyv1.CreateSchemaRequest{
			Name: name,
			Fields: []*tallyv1.SchemaRequestField{
				{Name: "amount", Type: tallyv1.FieldFormat_FIELD_FORMAT_NUMBER},
			},
		})
		_ = b.CreateSchema(reqBytes)
	}

	reqBytes, _ := proto.Marshal(&tallyv1.ListSchemasRequest{})
	outBytes := b.ListSchemas(reqBytes)

	var resp tallyv1.ListSchemasResponse
	if err := proto.Unmarshal(outBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal ListSchemas response: %v", err)
	}

	if resp.GetCode() != tallyv1.ResponseCode_RESPONSE_CODE_OK {
		t.Fatalf("ListSchemas() code = %v, want OK: %s", resp.GetCode(), resp.ErrorMessage)
	}

	if len(resp.GetSchemas()) != 2 {
		t.Errorf("expected 2 schemas, got %d", len(resp.GetSchemas()))
	}
}

func TestBridge_RecordAndListEntries(t *testing.T) {
	b := setupTestBridge(t)

	// Seed schema
	createReq := &tallyv1.CreateSchemaRequest{
		Name: "Hydration",
		Fields: []*tallyv1.SchemaRequestField{
			{Name: "ml", Type: tallyv1.FieldFormat_FIELD_FORMAT_INTEGER, Required: true},
		},
	}
	createBytes, _ := proto.Marshal(createReq)
	createOut := b.CreateSchema(createBytes)
	var createResp tallyv1.CreateSchemaResponse
	_ = proto.Unmarshal(createOut, &createResp)

	tallyID := createResp.GetTallyId()

	t.Run("RecordEntry success and validation failure", func(t *testing.T) {
		tests := []struct {
			name     string
			payload  string
			wantCode tallyv1.ResponseCode
		}{
			{
				name:     "valid JSON entry matching schema",
				payload:  `{"ml": 500}`,
				wantCode: tallyv1.ResponseCode_RESPONSE_CODE_OK,
			},
			{
				name:     "invalid JSON payload failing schema validation",
				payload:  `{"ml": "five hundred"}`,
				wantCode: tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				recordReqBytes, _ := proto.Marshal(&tallyv1.RecordEntryRequest{
					TallyId:     tallyID,
					PayloadJson: tt.payload,
				})

				recordOut := b.RecordEntry(recordReqBytes)
				var recordResp tallyv1.RecordEntryResponse
				if err := proto.Unmarshal(recordOut, &recordResp); err != nil {
					t.Fatalf("failed to unmarshal RecordEntryResponse: %v", err)
				}

				if recordResp.GetCode() != tt.wantCode {
					t.Errorf("RecordEntry() code = %v, want %v (msg: %s)", recordResp.GetCode(), tt.wantCode, recordResp.GetErrorMessage())
				}

				if tt.wantCode == tallyv1.ResponseCode_RESPONSE_CODE_OK && recordResp.GetEntryId() == "" {
					t.Errorf("expected EntryId to be populated on successful record")
				}
			})
		}
	})

	t.Run("ListEntries pagination", func(t *testing.T) {
		// Record a second entry so we have 2 total
		rec2Bytes, _ := proto.Marshal(&tallyv1.RecordEntryRequest{
			TallyId:     tallyID,
			PayloadJson: `{"ml": 750}`,
		})
		_ = b.RecordEntry(rec2Bytes)

		listReqBytes, _ := proto.Marshal(&tallyv1.ListEntriesRequest{
			TallyId: tallyID,
			Limit:   1,
			Offset:  0,
		})

		listOut := b.ListEntries(listReqBytes)
		var listResp tallyv1.ListEntriesResponse
		if err := proto.Unmarshal(listOut, &listResp); err != nil {
			t.Fatalf("failed to unmarshal ListEntriesResponse: %v", err)
		}

		if listResp.GetCode() != tallyv1.ResponseCode_RESPONSE_CODE_OK {
			t.Fatalf("ListEntries() code = %v, want OK", listResp.GetCode())
		}

		if listResp.GetTotalCount() != 2 {
			t.Errorf("expected TotalCount = 2, got %d", listResp.GetTotalCount())
		}

		if len(listResp.GetEntries()) != 1 {
			t.Errorf("expected 1 entry in page, got %d", len(listResp.GetEntries()))
		}

		if !listResp.GetHasMore() {
			t.Errorf("expected HasMore = true when page size < total count")
		}
	})
}
