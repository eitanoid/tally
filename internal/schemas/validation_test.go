package schemas

import (
	"testing"
)

func TestValidateDateTime(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   any
		wantErr bool
	}{
		{
			name:    "valid RFC3339 UTC",
			field:   "reading_started",
			value:   "2026-09-26T12:00:00Z",
			wantErr: false,
		},
		{
			name:    "valid RFC3339 with offset",
			field:   "reading_started",
			value:   "2026-09-26T13:00:00+01:00",
			wantErr: false,
		},
		{
			name:    "invalid format SQL style date",
			field:   "reading_started",
			value:   "2026-09-26 12:00:00",
			wantErr: true,
		},
		{
			name:    "invalid format date only",
			field:   "reading_started",
			value:   "2026-09-26",
			wantErr: true,
		},
		{
			name:    "invalid type integer instead of string",
			field:   "reading_started",
			value:   1727352000,
			wantErr: true,
		},
		{
			name:    "invalid type boolean",
			field:   "reading_started",
			value:   true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateTime(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTimestamp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   any
		wantErr bool
	}{
		{
			name:    "valid ISO date YYYY-MM-DD",
			field:   "target_date",
			value:   "2026-09-26",
			wantErr: false,
		},
		{
			name:    "invalid date format DD/MM/YYYY",
			field:   "target_date",
			value:   "26/09/2026",
			wantErr: true,
		},
		{
			name:    "invalid full timestamp instead of date only",
			field:   "target_date",
			value:   "2026-09-26T15:00:00Z",
			wantErr: true,
		},
		{
			name:    "invalid non-string type",
			field:   "target_date",
			value:   20260926,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDate(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTime(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   any
		wantErr bool
	}{
		{
			name:    "valid time with UTC offset Z",
			field:   "dose_time",
			value:   "14:30:00Z",
			wantErr: false,
		},
		{
			name:    "valid time with positive timezone offset",
			field:   "dose_time",
			value:   "15:30:00+01:00",
			wantErr: false,
		},
		{
			name:    "invalid time missing timezone offset",
			field:   "dose_time",
			value:   "14:30:00",
			wantErr: true,
		},
		{
			name:    "invalid time missing seconds",
			field:   "dose_time",
			value:   "14:30Z",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTime(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateISODuration(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   any
		wantErr bool
	}{
		{
			name:    "valid go duration seconds",
			field:   "brew_duration",
			value:   "28s",
			wantErr: false,
		},
		{
			name:    "valid Go duration minutes and seconds",
			field:   "brew_duration",
			value:   "30m12s",
			wantErr: false,
		},
		{
			name:    "valid Go duration hours",
			field:   "brew_duration",
			value:   "2h",
			wantErr: false,
		},
		{
			name:    "invalid ISO style duration string",
			field:   "brew_duration",
			value:   "PT15M30S",
			wantErr: true,
		},
		{
			name:    "invalid free-text duration",
			field:   "brew_duration",
			value:   "15 minutes",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGoDuration(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGoDuration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEntry_Integration(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"coffee_name": { "type": "string" },
			"dose_time": { "type": "string", "format": "time" },
			"brew_duration": { "type": "string", "format": "go-duration" },
			"logged_at": { "type": "string", "format": "date-time" }
		},
		"required": ["coffee_name", "dose_time", "brew_duration", "logged_at"],
		"additionalProperties": false
	}`

	tests := []struct {
		name    string
		payload string
		wantErr bool
	}{
		{
			name:    "valid full entry payload",
			payload: `{"coffee_name": "Ethiopia Yirgacheffe", "dose_time": "08:30:00Z", "brew_duration": "3m2s", "logged_at": "2026-09-26T08:30:00Z"}`,
			wantErr: false,
		},
		{
			name:    "invalid payload with non-Go duration (28s)",
			payload: `{"coffee_name": "Ethiopia Yirgacheffe", "dose_time": "08:30:00Z", "brew_duration": "PT28S", "logged_at": "2026-09-26T08:30:00Z"}`,
			wantErr: true,
		},
		{
			name:    "invalid payload with non-RFC3339 logged_at timestamp",
			payload: `{"coffee_name": "Ethiopia Yirgacheffe", "dose_time": "08:30:00Z", "brew_duration": "PT28S", "logged_at": "2026-09-26 08:30:00"}`,
			wantErr: true,
		},
		{
			name:    "invalid payload due to missing required field (dose_time)",
			payload: `{"coffee_name": "Ethiopia Yirgacheffe", "brew_duration": "PT28S", "logged_at": "2026-09-26T08:30:00Z"}`,
			wantErr: true,
		},
		{
			name:    "invalid payload due to unpermitted extra field (rating)",
			payload: `{"coffee_name": "Ethiopia Yirgacheffe", "dose_time": "08:30:00Z", "brew_duration": "PT28S", "logged_at": "2026-09-26T08:30:00Z", "rating": 5}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEntry(schemaJSON, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
