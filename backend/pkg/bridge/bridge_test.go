// Package bridge_test tests the foreign function interface (FFI) serialization and business logic boundary.
package bridge_test

import (
	"testing"

	"github.com/eitanoid/tally/generated/pb/tallyv1"
	"github.com/eitanoid/tally/pkg/bridge"
	"google.golang.org/protobuf/proto"
)

func setupTestBridge(t *testing.T) {
	t.Helper()
	bridge.Close()
	// Initialize Bridge with in-memory SQLite database
	err := bridge.New(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test bridge: %v", err)
	}
	return
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
					{
						Name:        "book form",
						Description: "Form of book",
						Type:        tallyv1.FieldFormat_FIELD_FORMAT_ONE_OF,
						Required:    false,
						EnumValues:  []string{"paperback", "hardback", "audiobook"},
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
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
		{
			name:         "handle malformed input bytes gracefully",
			corruptBytes: []byte("invalid-proto-bytes"),
			wantCode:     tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestBridge(t)

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

			outBytes := bridge.CreateSchema(inputBytes)

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
	setupTestBridge(t)

	// Seed schema
	createReq := &tallyv1.CreateSchemaRequest{
		Name:        "Coffee Log",
		Description: "Daily roasts",
		Fields: []*tallyv1.SchemaRequestField{
			{Name: "roast", Type: tallyv1.FieldFormat_FIELD_FORMAT_STRING, Required: true},
		},
	}
	createBytes, _ := proto.Marshal(createReq)
	createOut := bridge.CreateSchema(createBytes)
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
			outBytes := bridge.GetLatestSchema(reqBytes)

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
	setupTestBridge(t)

	// Seed 2 schemas
	for _, name := range []string{"Habits", "Finances"} {
		reqBytes, _ := proto.Marshal(&tallyv1.CreateSchemaRequest{
			Name: name,
			Fields: []*tallyv1.SchemaRequestField{
				{Name: "amount", Type: tallyv1.FieldFormat_FIELD_FORMAT_NUMBER},
			},
		})
		_ = bridge.CreateSchema(reqBytes)
	}

	reqBytes, _ := proto.Marshal(&tallyv1.ListSchemasRequest{})
	outBytes := bridge.ListSchemas(reqBytes)

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
	setupTestBridge(t)

	// Seed schema
	createReq := &tallyv1.CreateSchemaRequest{
		Name: "Hydration",
		Fields: []*tallyv1.SchemaRequestField{
			{Name: "ml", Type: tallyv1.FieldFormat_FIELD_FORMAT_INTEGER, Required: true},
		},
	}
	createBytes, _ := proto.Marshal(createReq)
	createOut := bridge.CreateSchema(createBytes)
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

				recordOut := bridge.RecordEntry(recordReqBytes)
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
		_ = bridge.RecordEntry(rec2Bytes)

		listReqBytes, _ := proto.Marshal(&tallyv1.ListEntriesRequest{
			TallyId: tallyID,
			Limit:   1,
			Offset:  0,
		})

		listOut := bridge.ListEntries(listReqBytes)
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

func TestBridge_UpdateEntry(t *testing.T) {
	setupTestBridge(t)

	// Seed schema and an entry to update
	createSchemaBytes, _ := proto.Marshal(&tallyv1.CreateSchemaRequest{
		Name: "Counter",
		Fields: []*tallyv1.SchemaRequestField{
			{Name: "count", Type: tallyv1.FieldFormat_FIELD_FORMAT_INTEGER, Required: true},
		},
	})
	createSchemaOut := bridge.CreateSchema(createSchemaBytes)
	var createSchemaResp tallyv1.CreateSchemaResponse
	_ = proto.Unmarshal(createSchemaOut, &createSchemaResp)
	tallyID := createSchemaResp.GetTallyId()

	recBytes, _ := proto.Marshal(&tallyv1.RecordEntryRequest{
		TallyId:     tallyID,
		PayloadJson: `{"count": 5}`,
	})
	recOut := bridge.RecordEntry(recBytes)
	var recResp tallyv1.RecordEntryResponse
	_ = proto.Unmarshal(recOut, &recResp)
	entryID := recResp.GetEntryId()

	tests := []struct {
		name         string
		req          *tallyv1.UpdateEntryRequest
		corruptBytes []byte
		wantCode     tallyv1.ResponseCode
	}{
		{
			name: "successfully update entry data",
			req: &tallyv1.UpdateEntryRequest{
				EntryId:   entryID,
				PatchData: `{"count": 10}`,
			},
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_OK,
		},
		{
			name: "reject update for non-existent entry",
			req: &tallyv1.UpdateEntryRequest{
				EntryId:   "non-existent-entry-id",
				PatchData: `{"count": 10}`,
			},
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
		{
			name: "reject patch payload failing schema validation",
			req: &tallyv1.UpdateEntryRequest{
				EntryId:   entryID,
				PatchData: `{"count": "not-an-integer"}`,
			},
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
		{
			name:         "handle malformed input bytes gracefully",
			corruptBytes: []byte("invalid-proto-bytes"),
			wantCode:     tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			outBytes := bridge.UpdateEntry(inputBytes)

			var resp tallyv1.UpdateEntryResponse
			if err := proto.Unmarshal(outBytes, &resp); err != nil {
				t.Fatalf("failed to unmarshal UpdateEntry response bytes: %v", err)
			}

			if resp.GetCode() != tt.wantCode {
				t.Errorf("UpdateEntry() code = %v, want %v (error_message: %s)", resp.GetCode(), tt.wantCode, resp.GetErrorMessage())
			}

			if tt.wantCode == tallyv1.ResponseCode_RESPONSE_CODE_OK {
				if resp.GetUpdatedEntry().GetEntryId() != entryID {
					t.Errorf("got EntryId %s, want %s", resp.GetUpdatedEntry().GetEntryId(), entryID)
				}
				if resp.GetUpdatedEntry().GetData() != `{"count":10}` {
					t.Errorf("got updated Data = %s, want %s", resp.GetUpdatedEntry().GetData(), `{"count":10}`)
				}
				if resp.GetUpdatedEntry().GetUpdatedAt() == nil {
					t.Errorf("expected UpdatedAt timestamp to be populated in updated proto entry")
				}
			}
		})
	}
}

func TestBridge_DeleteEntry(t *testing.T) {
	setupTestBridge(t)

	// Seed schema and an entry to delete
	createSchemaBytes, _ := proto.Marshal(&tallyv1.CreateSchemaRequest{
		Name: "Habits",
		Fields: []*tallyv1.SchemaRequestField{
			{Name: "done", Type: tallyv1.FieldFormat_FIELD_FORMAT_BOOLEAN, Required: true},
		},
	})
	createSchemaOut := bridge.CreateSchema(createSchemaBytes)
	var createSchemaResp tallyv1.CreateSchemaResponse
	_ = proto.Unmarshal(createSchemaOut, &createSchemaResp)

	recBytes, _ := proto.Marshal(&tallyv1.RecordEntryRequest{
		TallyId:     createSchemaResp.GetTallyId(),
		PayloadJson: `{"done": true}`,
	})
	recOut := bridge.RecordEntry(recBytes)
	var recResp tallyv1.RecordEntryResponse
	_ = proto.Unmarshal(recOut, &recResp)
	entryID := recResp.GetEntryId()

	tests := []struct {
		name         string
		req          *tallyv1.DeleteEntryRequest
		corruptBytes []byte
		wantCode     tallyv1.ResponseCode
	}{
		{
			name: "successfully soft-delete entry",
			req: &tallyv1.DeleteEntryRequest{
				EntryId: entryID,
			},
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_OK,
		},
		{
			name: "return error for non-existent entry",
			req: &tallyv1.DeleteEntryRequest{
				EntryId: "non-existent-entry-id",
			},
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
		{
			name:         "handle malformed input bytes gracefully",
			corruptBytes: []byte("invalid-proto-bytes"),
			wantCode:     tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			outBytes := bridge.DeleteEntry(inputBytes)

			var resp tallyv1.DeleteEntryResponse
			if err := proto.Unmarshal(outBytes, &resp); err != nil {
				t.Fatalf("failed to unmarshal DeleteEntry response bytes: %v", err)
			}

			if resp.GetCode() != tt.wantCode {
				t.Errorf("DeleteEntry() code = %v, want %v (error_message: %s)", resp.GetCode(), tt.wantCode, resp.GetErrorMessage())
			}
		})
	}
}

func TestBridge_DeleteTally(t *testing.T) {
	setupTestBridge(t)

	// Seed schema and an entry
	createSchemaBytes, _ := proto.Marshal(&tallyv1.CreateSchemaRequest{
		Name: "Water Log",
		Fields: []*tallyv1.SchemaRequestField{
			{Name: "ml", Type: tallyv1.FieldFormat_FIELD_FORMAT_INTEGER, Required: true},
		},
	})
	createSchemaOut := bridge.CreateSchema(createSchemaBytes)
	var createSchemaResp tallyv1.CreateSchemaResponse
	_ = proto.Unmarshal(createSchemaOut, &createSchemaResp)
	tallyID := createSchemaResp.GetTallyId()

	recBytes, _ := proto.Marshal(&tallyv1.RecordEntryRequest{
		TallyId:     tallyID,
		PayloadJson: `{"ml": 250}`,
	})
	_ = bridge.RecordEntry(recBytes)

	tests := []struct {
		name         string
		req          *tallyv1.DeleteTallyRequest
		corruptBytes []byte
		wantCode     tallyv1.ResponseCode
	}{
		{
			name: "successfully soft-delete tally",
			req: &tallyv1.DeleteTallyRequest{
				TallyId: tallyID,
			},
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_OK,
		},
		{
			name: "return error for non-existent tally",
			req: &tallyv1.DeleteTallyRequest{
				TallyId: "00000000-0000-0000-0000-000000000000",
			},
			wantCode: tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
		{
			name:         "handle malformed input bytes gracefully",
			corruptBytes: []byte("invalid-proto-bytes"),
			wantCode:     tallyv1.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			outBytes := bridge.DeleteTally(inputBytes)

			var resp tallyv1.DeleteTallyResponse
			if err := proto.Unmarshal(outBytes, &resp); err != nil {
				t.Fatalf("failed to unmarshal DeleteTally response bytes: %v", err)
			}

			if resp.GetCode() != tt.wantCode {
				t.Errorf("DeleteTally() code = %v, want %v (error_message: %s)", resp.GetCode(), tt.wantCode, resp.GetErrorMessage())
			}

			// Verify that soft-deleted tally schema is no longer accessible via GetLatestSchema
			if tt.wantCode == tallyv1.ResponseCode_RESPONSE_CODE_OK {
				getSchemaBytes, _ := proto.Marshal(&tallyv1.GetLatestSchemaRequest{TallyId: tallyID})
				getSchemaOut := bridge.GetLatestSchema(getSchemaBytes)
				var getSchemaResp tallyv1.GetLatestSchemaResponse
				_ = proto.Unmarshal(getSchemaOut, &getSchemaResp)

				if getSchemaResp.GetCode() == tallyv1.ResponseCode_RESPONSE_CODE_OK {
					t.Errorf("expected GetLatestSchema to fail for soft-deleted tally, but got OK")
				}
			}
		})
	}
}
