package schemas_test

import (
	"testing"

	"github.com/eitanoid/tally/internal/schemas"
	"github.com/google/jsonschema-go/jsonschema"
)

func TestBuildJSONSchema_FieldTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldName     string
		fieldDesc     string
		fieldType     schemas.SupportedType
		required      bool
		enumValues    []string
		expectedCheck func(t *testing.T, prop *jsonschema.Schema)
	}{
		{
			name:      "TypeString",
			fieldName: "Title",
			fieldDesc: "Book Title",
			fieldType: schemas.TypeString,
			required:  true,
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "string" {
					t.Errorf("expected type string, got %s", prop.Type)
				}
			},
		},
		{
			name:      "TypeInt",
			fieldName: "Age",
			fieldDesc: "User Age",
			fieldType: schemas.TypeInt,
			required:  false,
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "integer" {
					t.Errorf("expected type integer, got %s", prop.Type)
				}
			},
		},
		{
			name:      "TypeFloat",
			fieldName: "Price",
			fieldDesc: "Item Price",
			fieldType: schemas.TypeFloat,
			required:  true,
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "number" {
					t.Errorf("expected type number, got %s", prop.Type)
				}
			},
		},
		{
			name:      "TypeBool",
			fieldName: "Is Active",
			fieldDesc: "Account Active Status",
			fieldType: schemas.TypeBool,
			required:  false,
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "boolean" {
					t.Errorf("expected type boolean, got %s", prop.Type)
				}
			},
		},
		{
			name:      "TypeDateTime",
			fieldName: "Created At",
			fieldDesc: "Timestamp when record was created",
			fieldType: schemas.TypeDateTime,
			required:  true,
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "string" || prop.Format != "date-time" {
					t.Errorf("expected type string with format date-time, got type=%s format=%s", prop.Type, prop.Format)
				}
			},
		},
		{
			name:      "TypeDate",
			fieldName: "Birth Date",
			fieldDesc: "Date of Birth",
			fieldType: schemas.TypeDate,
			required:  false,
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "string" || prop.Format != "date" {
					t.Errorf("expected type string with format date, got type=%s format=%s", prop.Type, prop.Format)
				}
			},
		},
		{
			name:      "TypeTime",
			fieldName: "Alarm Time",
			fieldDesc: "Daily Alarm Time",
			fieldType: schemas.TypeTime,
			required:  false,
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "string" || prop.Format != "time" {
					t.Errorf("expected type string with format time, got type=%s format=%s", prop.Type, prop.Format)
				}
			},
		},
		{
			name:      "TypeDuration",
			fieldName: "Time Spent",
			fieldDesc: "Duration of task",
			fieldType: schemas.TypeDuration,
			required:  false,
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "string" || prop.Format != "go-duration" {
					t.Errorf("expected type string with format go-duration, got type=%s format=%s", prop.Type, prop.Format)
				}
			},
		},
		{
			name:       "TypeOneOf",
			fieldName:  "Status",
			fieldDesc:  "Current task status",
			fieldType:  schemas.TypeOneOf,
			required:   true,
			enumValues: []string{"todo", "in_progress", "done"},
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "string" {
					t.Errorf("expected type string, got %s", prop.Type)
				}
				if len(prop.Enum) != 3 {
					t.Fatalf("expected 3 enum values, got %d", len(prop.Enum))
				}
				if prop.Enum[0] != "todo" || prop.Enum[1] != "in_progress" || prop.Enum[2] != "done" {
					t.Errorf("unexpected enum values: %v", prop.Enum)
				}
			},
		},
		{
			name:       "TypeManyOf",
			fieldName:  "Tags",
			fieldDesc:  "Associated labels",
			fieldType:  schemas.TypeManyOf,
			required:   false,
			enumValues: []string{"go", "cli", "backend"},
			expectedCheck: func(t *testing.T, prop *jsonschema.Schema) {
				if prop.Type != "array" {
					t.Errorf("expected array type, got %s", prop.Type)
				}
				if !prop.UniqueItems {
					t.Error("expected UniqueItems to be true")
				}
				if prop.MinItems == nil || *prop.MinItems != 1 {
					t.Errorf("expected MinItems to be 1, got %v", prop.MinItems)
				}
				if prop.Items == nil {
					t.Fatal("expected Items schema to be initialized, got nil")
				}
				if prop.Items.Type != "string" {
					t.Errorf("expected item type string, got %s", prop.Items.Type)
				}
				if len(prop.Items.Enum) != 3 {
					t.Fatalf("expected 3 enum values in Items, got %d", len(prop.Items.Enum))
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := schemas.NewSchemaRequest("Test Schema", "Schema for unit tests")

			// Use appropriate fluent builder method based on type
			if tt.fieldType == schemas.TypeOneOf || tt.fieldType == schemas.TypeManyOf {
				req.WithEnumField(tt.fieldName, tt.fieldDesc, tt.fieldType, tt.required, tt.enumValues)
			} else {
				req.WithField(tt.fieldName, tt.fieldDesc, tt.fieldType, tt.required)
			}

			// Build JSON Schema
			schema, err := schemas.BuildJSONSchema(*req)
			if err != nil {
				t.Fatalf("unexpected error building schema: %v", err)
			}

			// Check global root property requirements
			expectedKey := schemas.Slugify(tt.fieldName)
			prop, exists := schema.Properties[expectedKey]
			if !exists {
				t.Fatalf("expected property key '%s' not found in generated schema", expectedKey)
			}

			if prop.Title != tt.fieldName {
				t.Errorf("expected Title '%s', got '%s'", tt.fieldName, prop.Title)
			}

			if prop.Description != tt.fieldDesc {
				t.Errorf("expected Description '%s', got '%s'", tt.fieldDesc, prop.Description)
			}

			// Verify required slice inclusion
			isListedAsRequired := false
			for _, reqKey := range schema.Required {
				if reqKey == expectedKey {
					isListedAsRequired = true
					break
				}
			}
			if tt.required && !isListedAsRequired {
				t.Errorf("expected field '%s' to be in required list, but it wasn't", expectedKey)
			} else if !tt.required && isListedAsRequired {
				t.Errorf("field '%s' was listed in required list unexpectedly", expectedKey)
			}

			// Perform field-type-specific checks
			tt.expectedCheck(t, prop)
		})
	}
}
