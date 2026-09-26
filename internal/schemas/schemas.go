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
	// would be nice to add units in the future to int and float
	TypeString   SupportedType = "string"
	TypeInt      SupportedType = "integer"
	TypeFloat    SupportedType = "number"
	TypeBool     SupportedType = "boolean"
	TypeDateTime SupportedType = "date-time"
	TypeDate     SupportedType = "date"
	TypeTime     SupportedType = "time"
	TypeDuration SupportedType = "duration"
)

var (
	ErrDuplicateField = errors.New("field name must be unique")
	ErrInvalidRequest = errors.New("request not valid")
)

func Valid(t SupportedType) bool {
	switch t {
	case TypeString, TypeInt, TypeFloat, TypeBool, TypeDateTime, TypeDate, TypeTime, TypeDuration:
		return true
	default:
		return false
	}
}

// FieldDefinition defines a single field that the user wants to track.
type FieldDefinition struct {
	Key         string        `json:"key"`   // eg. book_title (json safe key)
	Label       string        `json:"label"` // e.g. "Book Title" (user display label)
	Description string        `json:"description,omitempty"`
	Type        SupportedType `json:"type"`
	Required    bool          `json:"required"`
}

// SchemaRequest is the payload sent when a user creates a new tally.
type SchemaRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Fields      []FieldDefinition `json:"fields"`
	err         error             // fluent builder error
}

// TallySchema is what gets persisted in the DB.
type TallySchema struct {
	TallyID       string    `db:"tally_id" json:"tally_id"`
	Version       int       `json:"version" db:"version"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	JSONSchemaRaw string    `json:"json_schema"` // Raw JSON string stored in SQLite
	CreatedAt     time.Time `db:"created_at"  json:"created_at"`
}

// SchemaRef is the minimal identifier for fetching a specific schema version
type SchemaRef struct {
	TallyID string `db:"tally_id"`
	Version int    `json:"version"`
}

func (h *TallySchema) Ref() SchemaRef {
	return SchemaRef{
		TallyID: h.TallyID,
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
			Title:       field.Label,
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
		case TypeDateTime:
			propSchema.Type = "string"
			propSchema.Format = "date-time"
		case TypeTime:
			propSchema.Type = "string"
			propSchema.Format = "time"
		case TypeDate:
			propSchema.Type = "string"
			propSchema.Format = "date"
		case TypeDuration:
			propSchema.Type = "string"
			propSchema.Format = "go-duration" // it is not easy to use ISO 8601 durations in go
		default:
			return nil, fmt.Errorf("unsupported field type: %s", field.Type)
		}

		properties[field.Key] = propSchema

		if field.Required {
			requiredFields = append(requiredFields, field.Key)
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

func NewSchemaRequest(name, description string) *SchemaRequest {
	return &SchemaRequest{
		Name:        name,
		Description: description,
		Fields:      make([]FieldDefinition, 0),
	}
}

// Add field to schema and record the first error encountered
func (s *SchemaRequest) WithField(name, description string, typ SupportedType, required bool) *SchemaRequest {
	if s.err != nil {
		return s // Short-circuit if an error already occurred earlier in the chain
	}

	key := Slugify(name)
	if key == "" {
		s.err = fmt.Errorf("field name '%s' produced an empty key", name)
		return s
	}

	if !Valid(typ) {
		s.err = fmt.Errorf("invalid type '%s' for name '%s'", typ, name)
		return s
	}

	for _, existing := range s.Fields {
		if existing.Key == key {
			s.err = fmt.Errorf("duplicate field key '%s' (from name '%s')", key, name)
			return s
		}
	}

	s.Fields = append(s.Fields, FieldDefinition{
		Key:         key,
		Label:       name,
		Description: description,
		Type:        typ,
		Required:    required,
	})
	return s
}

// Build creates a new tally definition ready for SQLite insertion.
func (req *SchemaRequest) Build() (*TallySchema, error) {

	if req == nil {
		return nil, fmt.Errorf("failed to create tally: %w", ErrInvalidRequest)
	}
	if err := req.err; err != nil {
		return nil, err
	}

	schemaObj, err := BuildJSONSchema(*req)
	if err != nil {
		return nil, fmt.Errorf("failed to build a new json schema: %w", err)
	}

	// Serialize the schema struct to a JSON string for SQLite storage
	rawJSON, err := json.Marshal(schemaObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json schema: %w", err)
	}

	return &TallySchema{
		Name:          req.Name,
		Version:       1,
		TallyID:       ksuid.New().String(),
		Description:   req.Description,
		JSONSchemaRaw: string(rawJSON),
	}, nil
}

// ValidateEntry checks if an incoming dynamic payload matches a stored tally schema.
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

	// Validate against custom validation rules
	if err := runValidationRules(sch, pld); err != nil {
		return fmt.Errorf("custom validation failed: %w", err)
	}

	return nil
}
