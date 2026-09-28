// Package schemas provides domain types, JSON Schema generation, and validation logic.
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

// SupportedType represents all supported field formats for the caller (e.g. string, date-time, integer).
type SupportedType string

// Supported user defined types and formats for entries.
const (
	// TypeString represents an intent to store a String in a schema field.
	TypeString SupportedType = "string"
	// TypeInt represents an intent to store an Int in a schema field.
	TypeInt SupportedType = "integer"
	// TypeFloat represents an intent to store a Float in a schema field.
	TypeFloat SupportedType = "number"
	// TypeBool represents an intent to store a Bool in a schema field.
	TypeBool SupportedType = "boolean"
	// TypeDateTime represents an intent to store a DateTime in a schema field.
	TypeDateTime SupportedType = "date-time"
	// TypeDate represents an intent to store a Date in a schema field.
	TypeDate SupportedType = "date"
	// TypeTime represents an intent to store a Time in a schema field.
	TypeTime SupportedType = "time"
	// TypeDuration represents an intent to store a Duration in a schema field.
	TypeDuration SupportedType = "duration"
)

// Supported primitive types for JSON schema objects.
const (
	jsonString  = "string"
	jsonInteger = "integer"
	jsonNumber  = "number"
	jsonBoolean = "boolean"
)

// Supported standard and custom formats for JSON schema fields.
const (
	jsonFormatDateTime = "date-time"
	jsonFormatTime     = "time"
	jsonFormatDate     = "date"
	// Custom format for Go duration strings (e.g. "2h3m").
	jsonFormatDuration = "go-duration"
)

var (
	// ErrDuplicateField is raised when a field name is duplicated.
	ErrDuplicateField = errors.New("field name must be unique")
	// ErrInvalidRequest is raised when a SchemaRequest is not valid
	ErrInvalidRequest = errors.New("request not valid")
)

// Valid validates a SupportedType
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

// Ref returns the unique identifier for a schema.
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
			propSchema.Type = jsonString
		case TypeInt:
			propSchema.Type = jsonInteger
		case TypeFloat:
			propSchema.Type = jsonNumber
		case TypeBool:
			propSchema.Type = jsonBoolean
		case TypeDateTime:
			propSchema.Type = jsonString
			propSchema.Format = jsonFormatDateTime
		case TypeTime:
			propSchema.Type = jsonString
			propSchema.Format = jsonFormatTime
		case TypeDate:
			propSchema.Type = jsonString
			propSchema.Format = jsonFormatDate
		case TypeDuration:
			propSchema.Type = jsonString
			propSchema.Format = jsonFormatDuration
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

// NewSchemaRequest creates a new SchemaRequest object.
func NewSchemaRequest(name, description string) *SchemaRequest {
	return &SchemaRequest{
		Name:        name,
		Description: description,
		Fields:      make([]FieldDefinition, 0),
	}
}

// WithField adds a field to a SchemaRequest and records the first error encountered.
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
func (s *SchemaRequest) Build() (*TallySchema, error) {

	if s == nil {
		return nil, fmt.Errorf("failed to create tally: %w", ErrInvalidRequest)
	}
	if err := s.err; err != nil {
		return nil, err
	}

	schemaObj, err := BuildJSONSchema(*s)
	if err != nil {
		return nil, fmt.Errorf("failed to build a new json schema: %w", err)
	}

	// Serialize the schema struct to a JSON string for SQLite storage
	rawJSON, err := json.Marshal(schemaObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json schema: %w", err)
	}

	return &TallySchema{
		Name:          s.Name,
		Version:       1,
		TallyID:       ksuid.New().String(),
		Description:   s.Description,
		JSONSchemaRaw: string(rawJSON),
	}, nil
}

// ValidateJSONData checks if an incoming dynamic payload matches a stored tally schema.
func ValidateJSONData(jsonSchema string, payload string) error {
	var pld map[string]any
	if err := json.Unmarshal([]byte(payload), &pld); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	var sch jsonschema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &sch); err != nil {
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

// ValidatePatch ensures the patch payload itself is valid JSON.
func ValidatePatch(jsonPatch []byte) error {
	var pld any
	if err := json.Unmarshal(jsonPatch, &pld); err != nil {
		return fmt.Errorf("invalid patch JSON payload: %w", err)
	}

	// JSON Merge Patch (RFC 7396) root must be a JSON object
	if _, ok := pld.(map[string]any); !ok {
		return fmt.Errorf("patch payload must be a JSON object")
	}

	return nil
}
