// Package entries_test test the entries package.
package entries_test

import (
	"testing"

	"github.com/eitanoid/tally/internal/entries"
	"github.com/eitanoid/tally/internal/schemas"
)

func TestCreateTallyEntry(t *testing.T) {
	validSchema := schemas.TallySchema{
		TallyID:       "550e8400-e29b-41d4-a716-446655440000",
		Version:       1,
		JSONSchemaRaw: `{"type":"object","properties":{"count":{"type":"integer"}},"required":["count"]}`,
	}

	tests := []struct {
		name    string
		schema  schemas.TallySchema
		data    string
		wantErr bool
	}{
		{
			name:    "valid entry payload",
			schema:  validSchema,
			data:    `{"count": 5}`,
			wantErr: false,
		},
		{
			name:    "invalid entry payload according to schema",
			schema:  validSchema,
			data:    `{"count": "not an integer"}`,
			wantErr: true,
		},
		{
			name:    "malformed JSON data",
			schema:  validSchema,
			data:    `{invalid-json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := entries.CreateTallyEntry(tt.schema, tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateTallyEntry() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if got.ID == "" {
					t.Errorf("expected KSUID ID to be generated, got empty string")
				}
				if got.TallyID != tt.schema.TallyID {
					t.Errorf("got TallyID %s, want %s", got.TallyID, tt.schema.TallyID)
				}
				if got.SchemaVersion != tt.schema.Version {
					t.Errorf("got SchemaVersion %d, want %d", got.SchemaVersion, tt.schema.Version)
				}
			}
		})
	}
}

func TestTallyEntry_SchemaRef(t *testing.T) {
	entry := entries.TallyEntry{
		TallyID:       "550e8400-e29b-41d4-a716-446655440000",
		SchemaVersion: 2,
	}

	ref := entry.SchemaRef()
	if ref.TallyID != entry.TallyID || ref.Version != entry.SchemaVersion {
		t.Errorf("SchemaRef() = %+v, want TallyID=%s Version=%d", ref, entry.TallyID, entry.SchemaVersion)
	}
}
