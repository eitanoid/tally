package entries

import (
	"fmt"
	"time"

	"github.com/eitanoid/tally/internal/schemas"
	"github.com/segmentio/ksuid"
)

type TallyEntry struct {
	ID            string    `db:"id" json:"id"`
	TallyID       string    `db:"tally_id" json:"tally_id"`
	SchemaVersion int       `json:"schema_version" db:"schema_version"`
	Data          string    `db:"data" json:"data"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

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

func (e *TallyEntry) SchemaRef() schemas.SchemaRef {
	return schemas.SchemaRef{
		TallyID: e.TallyID,
		Version: e.SchemaVersion,
	}
}
