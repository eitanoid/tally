package schemas

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	// "reflect"
)

/*
 example habit?
 what if we want to add constraits eg. Dosage is of type weight
 {
 	Name:
	Description:
	Schema: {
	  -	Name: Medicine
		Desciption: Medicine that you take
		Type: String
	  -	Name: Dose
		Desciption: How much medication you took
		Type: Int
	  -	Name: TakenAt
		Desciption: When you took your last dose
		Type: Timestamp
	}
 }

*/

type SupportedType string

const (
	TypeString    SupportedType = "string"
	TypeInt       SupportedType = "integer" // support units in the future?
	TypeFloat     SupportedType = "number"
	TypeBool      SupportedType = "boolean"
	TypeTimestamp SupportedType = "timestamp"
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
	Name          string `json:"name"`
	Description   string `json:"description"`
	JSONSchemaRaw string `json:"json_schema"` // Raw JSON string stored in SQLite
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
		Type:                 "object",
		Description:          req.Description,
		Properties:           properties,
		Required:             requiredFields,
		AdditionalProperties: &jsonschema.Schema{}, // Rejects unregistered dynamic fields
	}

	return schema, nil
}

// CreateHabitSchema creates a new habit definition ready for SQLite insertion.
func CreateHabitSchema(req SchemaRequest) (*HabitSchema, error) {
	schemaObj, err := BuildJSONSchema(req)
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
		Description:   req.Description,
		JSONSchemaRaw: string(rawJSON),
	}, nil
}

// ValidateEntry checks if an incoming dynamic payload matches a stored habit schema.
func ValidateEntry(rawSchemaJSON string, payload map[string]any) error {
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
	if err := resolved.Validate(payload); err != nil {
		return fmt.Errorf("entry validation failed: %w", err)
	}

	return nil
}
