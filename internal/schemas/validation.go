package schemas

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
)

var (
	// KeyPattern ensures field keys are safe for SQLite JSON paths & JS property access
	KeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// NonAlphaNumeric regex for auto-slugifying labels to keys
	NonAlphaNumeric = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
)

// Slugify converts a user label like "Book Title!" into a clean key "book_title"
func Slugify(label string) string {
	clean := strings.TrimSpace(strings.ToLower(label))
	clean = strings.ReplaceAll(clean, " ", "_")
	clean = NonAlphaNumeric.ReplaceAllString(clean, "")
	return clean
}

// walk properties and validate with custom rules
var propertyValidator = map[string]func(string, any) error{
	"date-time":   validateDateTime,
	"time":        validateTime,
	"date":        validateDate,
	"go-duration": validateGoDuration,
}

// validateDateTime enforces full RFC3339 timestamps (e.g. 2026-09-26T15:00:00Z)
func validateDateTime(field string, val any) error {
	strVal, ok := val.(string)
	if !ok {
		return fmt.Errorf("field '%s' must be a string (got %T)", field, val)
	}

	// Enforce strict RFC3339 / ISO-8601 parsing
	if _, err := time.Parse(time.RFC3339, strVal); err != nil {
		return fmt.Errorf("invalid timestamp format '%s' (expected RFC3339, e.g. 2026-09-26T12:00:00Z)", strVal)
	}
	return nil
}

// validateDate enforces ISO-8601 full dates without time (e.g. 2026-09-26)
func validateDate(field string, val any) error {
	strVal, ok := val.(string)
	if !ok {
		return fmt.Errorf("field '%s' must be a string", field)
	}
	const dateLayout = "2006-01-02"
	if _, err := time.Parse(dateLayout, strVal); err != nil {
		return fmt.Errorf("field '%s' has invalid date format '%s' (expected 'YYYY-MM-DD', e.g. '2026-09-26')", field, strVal)
	}
	return nil
}

// validateTime enforces ISO-8601 time-of-day with timezone offset (e.g. 14:30:00Z or 15:30:00+01:00)
func validateTime(field string, val any) error {
	strVal, ok := val.(string)
	if !ok {
		return fmt.Errorf("field '%s' must be a string", field)
	}
	const timeLayout = "15:04:05Z07:00"
	if _, err := time.Parse(timeLayout, strVal); err != nil {
		return fmt.Errorf("field '%s' has invalid time format '%s' (expected 'HH:MM:SSZ', e.g. '14:30:00Z' or '15:30:00+01:00')", field, strVal)
	}
	return nil
}

// validateGoDuration enforces golang style durations (e.g. 30m or 1h30s)
func validateGoDuration(field string, val any) error {
	strVal, ok := val.(string)
	if !ok {
		return fmt.Errorf("field '%s' must be a string (got %T)", field, val)
	}
	d, err := time.ParseDuration(strVal)
	if err != nil {
		return fmt.Errorf("field '%s' has invalid duration '%s' (expected e.g. '15m30s', '28s', '2h'): %w", field, strVal, err)
	}
	if d <= 0 {
		return fmt.Errorf("field '%s' duration must be greater than zero", field)
	}
	return nil
}

// runValidationRules runs custom format assertions on payload data
func runValidationRules(sch jsonschema.Schema, payload map[string]any) error {
	for propName, propSchema := range sch.Properties {
		if propSchema == nil {
			continue
		}
		// run custom validation
		if validation, ok := propertyValidator[propSchema.Format]; ok {
			val, exists := payload[propName]
			if !exists || val == nil {
				continue // Handled by schema 'required' validation if mandatory
			}
			if err := validation(propName, val); err != nil {
				return err
			}
		}
	}
	return nil
}
