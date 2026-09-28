// Package entries defines domain models and constructor logic for user tally entries.
package entries

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/eitanoid/tally/internal/schemas"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/segmentio/ksuid"
)

var (
	ErrNotFound     = errors.New("entry not found")
	ErrInvalidData  = errors.New("data does not comply with schema")
	ErrInvalidPatch = errors.New("invalid patch")
)

// TallyEntry represents a single recorded data point tied to a specific TallySchema version.
type TallyEntry struct {
	ID            string    `db:"id" json:"id"`
	TallyID       string    `db:"tally_id" json:"tally_id"`
	SchemaVersion int       `json:"schema_version" db:"schema_version"`
	Data          string    `db:"data" json:"data"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// CreateTallyEntry creates a TallyEntry against a TallySchema and validates the data.
func CreateTallyEntry(schema schemas.TallySchema, data string) (TallyEntry, error) {
	// validate against schema
	if err := schemas.ValidateEntry(schema.JSONSchemaRaw, data); err != nil {
		return TallyEntry{}, fmt.Errorf("failed to validate entry data: %w", err)
	}
	return TallyEntry{
		ID:            ksuid.New().String(),
		TallyID:       schema.TallyID,
		SchemaVersion: schema.Version,
		Data:          data,
	}, nil
}

// SchemaRef returns the unique SchemaRef for the TallyEntry.
func (e *TallyEntry) SchemaRef() schemas.SchemaRef {
	return schemas.SchemaRef{
		TallyID: e.TallyID,
		Version: e.SchemaVersion,
	}
}

// UpdateEntry acepts a JSON patch for an existing Entry with a schema.
func UpdateEntry(existingEntry *TallyEntry, rawPatchData []byte, schemaJSON []byte) (*TallyEntry, error) {
	if err := schemas.ValidatePatch(rawPatchData, schemaJSON); err != nil {
		return nil, fmt.Errorf("invalid patch data: %w", err)
	}

	originalJSON, err := json.Marshal(existingEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal original entry: %w", err)
	}

	patchedJSON, err := jsonpatch.MergePatch(originalJSON, rawPatchData)
	if err != nil {
		return nil, fmt.Errorf("failed to apply merge patch: %w", err)
	}

	// ensure patched data still complies to the schema
	if err := schemas.ValidateEntry(string(patchedJSON), string(schemaJSON)); err != nil {
		return nil, fmt.Errorf("patched object violates full schema: %w", err)
	}

	var updatedEntry TallyEntry
	if err := json.Unmarshal(patchedJSON, &updatedEntry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal patched entry: %w", err)
	}

	return &updatedEntry, nil
}
