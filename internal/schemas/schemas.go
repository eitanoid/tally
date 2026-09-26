package schemas

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/segmentio/ksuid"
)

// decisions to be made:
//
// naming conventions for fields? allow spaces? repalce spaces with - or _?
// default fields like `added` might be worth making an uncommon name like `__added__` and interpreting that later since `added` could be a common name for a field
// add logged at to db entry instead of schema

type SupportedType string

const (
	TypeString    SupportedType = "string"
	TypeInt       SupportedType = "integer" // support units in the future?
	TypeFloat     SupportedType = "number"
	TypeBool      SupportedType = "boolean"
	TypeTimestamp SupportedType = "timestamp"
)

var (
	ErrDuplicateField = errors.New("field name must be unique")
	ErrInvalidRequest = errors.New("request not valid")
)

func Validate(s SupportedType) (SupportedType, error) {
	switch s {
	case TypeString, TypeInt, TypeFloat, TypeBool:
		return s, nil
	default:
		return s, errors.New("Invalid type")
	}
}

// SchemaKey defines a single field that the user wants to track.
type SchemaKey struct {
	KeyName     string        `json:"key_name"`
	Description string        `json:"description,omitempty"`
	Type        SupportedType `json:"type"`
	Required    bool          `json:"required"`
}

// SchemaRequest is the payload sent when a user creates a new habit.
type SchemaRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Fields      []SchemaKey `json:"fields"`
}

// HabitSchema is what gets persisted in the DB.
type HabitSchema struct {
	HabitID       string    `db:"habit_id" json:"habit_id"`
	Version       int       `json:"version" db:"version"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	JSONSchemaRaw string    `json:"json_schema"` // Raw JSON string stored in SQLite
	CreatedAt     time.Time `db:"created_at"  json:"created_at"`
}

// SchemaRef is the minimal identifier for fetching a specific schema version
type SchemaRef struct {
	HabitID string `db:"habit_id"`
	Version int    `json:"version"`
}

func (h *HabitSchema) Ref() SchemaRef {
	return SchemaRef{
		HabitID: h.HabitID,
		Version: h.Version,
	}
}

// BuildJSONSchema converts a user's field definitions into a valid JSON Schema object.
func BuildJSONSchema(req SchemaRequest) (*jsonschema.Schema, error) {
	properties := make(map[string]*jsonschema.Schema)
	var requiredFields []string

	for _, field := range req.Fields {
		propSchema := &jsonschema.Schema{
			Description: field.Description,
		}

		switch field.Type {
		case TypeString:
			propSchema.Type = "string"
		case TypeInt:
			propSchema.Type = "integer"
		case TypeFloat:
			propSchema.Type = "number"
		case TypeBool:
			propSchema.Type = "boolean"
		case TypeTimestamp:
			propSchema.Type = "string"
			propSchema.Format = "date-time"
		default:
			return nil, fmt.Errorf("unsupported field type: %s", field.Type)
		}

		properties[field.KeyName] = propSchema

		if field.Required {
			requiredFields = append(requiredFields, field.KeyName)
		}
	}

	// Construct the root object schema
	schema := &jsonschema.Schema{
		Type:        "object",
		Description: req.Description,
		Properties:  properties,
		Required:    requiredFields,
		AdditionalProperties: &jsonschema.Schema{
			Not: &jsonschema.Schema{},
		}, // Rejects unregistered dynamic fields
	}

	return schema, nil
}

// Validate schema request doesn't contain duplicate key names
func (s *SchemaRequest) Validate() error {
	keys := make(map[string]struct{}, len(s.Fields))
	for _, field := range s.Fields {
		if _, exists := keys[field.KeyName]; exists {
			return ErrDuplicateField
		}
	}
	return nil
}

func NewSchemaRequest(name, description string) *SchemaRequest {
	return &SchemaRequest{
		Name:        name,
		Description: description,
	}
}

// Add field to schema
func (s *SchemaRequest) WithField(name, descripton string, typ SupportedType, required bool) *SchemaRequest {
	s.Fields = append(s.Fields, SchemaKey{
		KeyName:     name,
		Description: descripton,
		Type:        typ,
		Required:    required,
	})
	return s
}

// CreateHabitSchema creates a new habit definition ready for SQLite insertion.
func (req *SchemaRequest) Create() (*HabitSchema, error) {

	if req == nil || req.Fields == nil {
		return nil, fmt.Errorf("failed to create habit: %w", ErrInvalidRequest)
	}

	err := req.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to create habit: %w", err)
	}

	schemaObj, err := BuildJSONSchema(*req)
	if err != nil {
		return nil, fmt.Errorf("failed to build a new json shcema: %w", err)
	}

	// Serialize the schema struct to a JSON string for SQLite storage
	rawJSON, err := json.Marshal(schemaObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json schema: %w", err)
	}

	return &HabitSchema{
		Name:          req.Name,
		Version:       1,
		HabitID:       ksuid.New().String(),
		Description:   req.Description,
		JSONSchemaRaw: string(rawJSON),
	}, nil
}

// ValidateEntry checks if an incoming dynamic payload matches a stored habit schema.
func ValidateEntry(rawSchemaJSON string, payload string) error {
	var pld map[string]any
	if err := json.Unmarshal([]byte(payload), &pld); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	var sch jsonschema.Schema
	if err := json.Unmarshal([]byte(rawSchemaJSON), &sch); err != nil {
		return fmt.Errorf("invalid stored schema: %w", err)
	}

	// Resolve references and prepare validator
	resolved, err := sch.Resolve(nil)
	if err != nil {
		return fmt.Errorf("schema resolution failed: %w", err)
	}

	// Validate incoming map against the schema
	if err := resolved.Validate(pld); err != nil {
		return fmt.Errorf("entry validation failed: %w", err)
	}

	return nil
}
