package main

import (
	"path/filepath"
	"testing"

	"github.com/eitanoid/tally/generated/pb/tallyv1"
	"github.com/eitanoid/tally/pkg/bridge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

const initialEntryData = `{
	"activity": "Run",
	"count": 3,
	"distance": 5.25,
	"completed": true,
	"recorded_at": "2026-10-02T08:30:00Z",
	"day": "2026-10-02",
	"start_time": "08:30:00Z",
	"duration": "35m",
	"intensity": "moderate",
	"tags": ["fitness", "morning"]
}`

func TestE2ETallyEntryLifecycle(t *testing.T) {
	tests := []struct {
		name     string
		patch    string
		wantData string
	}{
		{
			name:  "edit text field",
			patch: `{"activity":"Swim"}`,
			wantData: `{
				"activity":"Swim","count":3,"distance":5.25,"completed":true,
				"recorded_at":"2026-10-02T08:30:00Z","day":"2026-10-02",
				"start_time":"08:30:00Z","duration":"35m","intensity":"moderate",
				"tags":["fitness","morning"]
			}`,
		},
		{
			name:  "edit numeric and enum fields",
			patch: `{"count":4,"distance":8.5,"intensity":"high","tags":["fitness","training"]}`,
			wantData: `{
				"activity":"Run","count":4,"distance":8.5,"completed":true,
				"recorded_at":"2026-10-02T08:30:00Z","day":"2026-10-02",
				"start_time":"08:30:00Z","duration":"35m","intensity":"high",
				"tags":["fitness","training"]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, bridge.Close())
			require.NoError(t, bridge.New(filepath.Join(t.TempDir(), "e2e.sqlite")))
			t.Cleanup(func() { require.NoError(t, bridge.Close()) })

			createSchemaBytes, err := proto.Marshal(allSupportedFieldsRequest())
			require.NoError(t, err)
			var createSchemaResp tallyv1.CreateSchemaResponse
			require.NoError(t, proto.Unmarshal(bridge.CreateSchema(createSchemaBytes), &createSchemaResp))
			require.Equal(t, tallyv1.ResponseCode_RESPONSE_CODE_OK, createSchemaResp.GetCode(), createSchemaResp.GetErrorMessage())
			require.NotEmpty(t, createSchemaResp.GetTallyId())
			assert.Equal(t, int32(1), createSchemaResp.GetSchemaVersion())

			recordEntryBytes, err := proto.Marshal(&tallyv1.RecordEntryRequest{
				TallyId:     createSchemaResp.GetTallyId(),
				PayloadJson: initialEntryData,
			})
			require.NoError(t, err)
			var recordEntryResp tallyv1.RecordEntryResponse
			require.NoError(t, proto.Unmarshal(bridge.RecordEntry(recordEntryBytes), &recordEntryResp))
			require.Equal(t, tallyv1.ResponseCode_RESPONSE_CODE_OK, recordEntryResp.GetCode(), recordEntryResp.GetErrorMessage())
			require.NotEmpty(t, recordEntryResp.GetEntryId())

			listed := listEntries(t, createSchemaResp.GetTallyId())
			require.Equal(t, int32(1), listed.GetTotalCount())
			require.Len(t, listed.GetEntries(), 1)
			assert.Equal(t, recordEntryResp.GetEntryId(), listed.GetEntries()[0].GetEntryId())
			assert.JSONEq(t, initialEntryData, listed.GetEntries()[0].GetData())

			updateBytes, err := proto.Marshal(&tallyv1.UpdateEntryRequest{
				EntryId:   recordEntryResp.GetEntryId(),
				PatchData: tt.patch,
			})
			require.NoError(t, err)
			var updateResp tallyv1.UpdateEntryResponse
			require.NoError(t, proto.Unmarshal(bridge.UpdateEntry(updateBytes), &updateResp))
			require.Equal(t, tallyv1.ResponseCode_RESPONSE_CODE_OK, updateResp.GetCode(), updateResp.GetErrorMessage())
			require.NotNil(t, updateResp.GetUpdatedEntry())
			assert.Equal(t, recordEntryResp.GetEntryId(), updateResp.GetUpdatedEntry().GetEntryId())
			assert.JSONEq(t, tt.wantData, updateResp.GetUpdatedEntry().GetData())

			listed = listEntries(t, createSchemaResp.GetTallyId())
			require.Equal(t, int32(1), listed.GetTotalCount())
			assert.JSONEq(t, tt.wantData, listed.GetEntries()[0].GetData())

			deleteBytes, err := proto.Marshal(&tallyv1.DeleteEntryRequest{EntryId: recordEntryResp.GetEntryId()})
			require.NoError(t, err)
			var deleteResp tallyv1.DeleteEntryResponse
			require.NoError(t, proto.Unmarshal(bridge.DeleteEntry(deleteBytes), &deleteResp))
			require.Equal(t, tallyv1.ResponseCode_RESPONSE_CODE_OK, deleteResp.GetCode(), deleteResp.GetErrorMessage())

			listed = listEntries(t, createSchemaResp.GetTallyId())
			assert.Zero(t, listed.GetTotalCount())
			assert.Empty(t, listed.GetEntries())
		})
	}
}

func listEntries(t *testing.T, tallyID string) *tallyv1.ListEntriesResponse {
	t.Helper()

	requestBytes, err := proto.Marshal(&tallyv1.ListEntriesRequest{
		TallyId: tallyID,
		Limit:   20,
	})
	require.NoError(t, err)

	var response tallyv1.ListEntriesResponse
	require.NoError(t, proto.Unmarshal(bridge.ListEntries(requestBytes), &response))
	require.Equal(t, tallyv1.ResponseCode_RESPONSE_CODE_OK, response.GetCode(), response.GetErrorMessage())
	return &response
}

func allSupportedFieldsRequest() *tallyv1.CreateSchemaRequest {
	return &tallyv1.CreateSchemaRequest{
		Name:        "Full field coverage",
		Description: "Exercise tally covering every supported field type",
		Fields: []*tallyv1.SchemaRequestField{
			{Name: "activity", Description: "Activity name", Type: tallyv1.FieldFormat_FIELD_FORMAT_STRING, Required: true},
			{Name: "count", Description: "Repetition count", Type: tallyv1.FieldFormat_FIELD_FORMAT_INTEGER, Required: true},
			{Name: "distance", Description: "Distance", Type: tallyv1.FieldFormat_FIELD_FORMAT_NUMBER, Required: true},
			{Name: "completed", Description: "Whether it was completed", Type: tallyv1.FieldFormat_FIELD_FORMAT_BOOLEAN, Required: true},
			{Name: "recorded_at", Description: "Record timestamp", Type: tallyv1.FieldFormat_FIELD_FORMAT_DATE_TIME, Required: true},
			{Name: "day", Description: "Activity date", Type: tallyv1.FieldFormat_FIELD_FORMAT_DATE, Required: true},
			{Name: "start_time", Description: "Start time", Type: tallyv1.FieldFormat_FIELD_FORMAT_TIME, Required: true},
			{Name: "duration", Description: "Activity duration", Type: tallyv1.FieldFormat_FIELD_FORMAT_DURATION, Required: true},
			{Name: "intensity", Description: "Effort level", Type: tallyv1.FieldFormat_FIELD_FORMAT_ONE_OF, Required: true, EnumValues: []string{"low", "moderate", "high"}},
			{Name: "tags", Description: "Activity tags", Type: tallyv1.FieldFormat_FIELD_FORMAT_MANY_OF, Required: true, EnumValues: []string{"fitness", "morning", "training"}},
		},
	}
}
