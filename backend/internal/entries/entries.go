// Package entries defines domain models and constructor logic for user tally entries.
package entries

import (
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
	if err := schemas.ValidateJSONData(schema.JSONSchemaRaw, data); err != nil {
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

// UpdateEntry accepts a JSON patch for an existing Entry with a schema.
func UpdateEntry(existingEntry *TallyEntry, jsonPatchData []byte, jsonSchema []byte) (*TallyEntry, error) {
	if existingEntry == nil {
		return nil, ErrNotFound
	}

	// validate Patch is a valid JSON
	if err := schemas.ValidatePatch(jsonPatchData); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPatch, err)
	}

	// apply merge-patch
	patchedJSON, err := jsonpatch.MergePatch([]byte(existingEntry.Data), jsonPatchData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPatch, err)
	}

	// validate merged object satisfies the schema
	if err := schemas.ValidateJSONData(string(jsonSchema), string(patchedJSON)); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidData, err)
	}

	updatedEntry := *existingEntry
	updatedEntry.Data = string(patchedJSON)

	return &updatedEntry, nil
}
