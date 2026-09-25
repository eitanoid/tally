package entries

import (
	"fmt"
	"time"

	"github.com/eitanoid/habit-tracker/internal/schemas"
	"github.com/segmentio/ksuid"
)

type HabitEntry struct {
	ID            string    `db:"id" json:"id"`
	HabitID       string    `db:"habit_id" json:"habit_id"`
	SchemaVersion int       `json:"schema_version" db:"schema_version"`
	Data          string    `db:"data" json:"data"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

func CreateHabitEntry(schema schemas.HabitSchema, data string) (HabitEntry, error) {

	// validate against schema
	if err := schemas.ValidateEntry(schema.JSONSchemaRaw, data); err != nil {
		return HabitEntry{}, fmt.Errorf("failed to validate entry data: %w", err)
	}
	return HabitEntry{
		ID:            ksuid.New().String(),
		HabitID:       schema.HabitID,
		SchemaVersion: schema.Version,
		Data:          data,
	}, nil
}

func (e *HabitEntry) SchemaRef() schemas.SchemaRef {
	return schemas.SchemaRef{
		HabitID: e.HabitID,
		Version: e.SchemaVersion,
	}
}
