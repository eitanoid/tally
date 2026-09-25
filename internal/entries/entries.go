package entries

import (
	"time"

	"github.com/eitanoid/habit-tracker/internal/schemas"
	"github.com/segmentio/ksuid"
)

type EntryRequest struct {
}

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
	// instanciate schema
	return HabitEntry{
		ID:            ksuid.New().String(),
		HabitID:       schema.HabitID,
		SchemaVersion: schema.Version,
	}, nil
}

func (e *HabitEntry) SchemaRef() schemas.SchemaRef {
	return schemas.SchemaRef{
		HabitID: e.HabitID,
		Version: e.SchemaVersion,
	}
}
