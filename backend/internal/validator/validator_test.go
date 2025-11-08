package validator_test

import (
	"testing"

	"github.com/johndennehy101/note-taking-web-app/backend/internal/validator"
)

func TestNew(t *testing.T) {
	v := validator.New()

	if v == nil {
		t.Fatal("expected New() to return a non-nil Validator")
	}

	if v.Errors == nil {
		t.Error("expected Errors map to be initialized")
	}

	if len(v.Errors) != 0 {
		t.Errorf("expected empty Errors map, got %d errors", len(v.Errors))
	}
}

func TestValidator_Valid(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *validator.Validator
		expectedValid bool
	}{
		{
			name:          "no errors",
			setup:         validator.New,
			expectedValid: true,
		},
		{
			name: "one error",
			setup: func() *validator.Validator {
				v := validator.New()
				v.AddError("field", "error message")
				return v
			},
			expectedValid: false,
		},
		{
			name: "multiple errors",
			setup: func() *validator.Validator {
				v := validator.New()
				v.AddError("field1", "error1")
				v.AddError("field2", "error2")
				return v
			},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			if v.Valid() != tt.expectedValid {
				t.Errorf("expected Valid() to return %v, got %v", tt.expectedValid, v.Valid())
			}
		})
	}
}

func TestValidator_AddError(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *validator.Validator
		key            string
		message        string
		expectedErrors int
		expectedMsg    string
	}{
		{
			name:           "add first error",
			setup:          validator.New,
			key:            "field",
			message:        "error message",
			expectedErrors: 1,
			expectedMsg:    "error message",
		},
		{
			name: "add error to existing validator",
			setup: func() *validator.Validator {
				v := validator.New()
				v.AddError("field1", "error1")
				return v
			},
			key:            "field2",
			message:        "error2",
			expectedErrors: 2,
			expectedMsg:    "error2",
		},
		{
			name: "duplicate key does not overwrite",
			setup: func() *validator.Validator {
				v := validator.New()
				v.AddError("field", "original message")
				return v
			},
			key:            "field",
			message:        "new message",
			expectedErrors: 1,
			expectedMsg:    "original message", // Should keep original
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			v.AddError(tt.key, tt.message)

			if len(v.Errors) != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d", tt.expectedErrors, len(v.Errors))
			}

			if msg, exists := v.Errors[tt.key]; exists {
				if msg != tt.expectedMsg {
					t.Errorf("expected message %q, got %q", tt.expectedMsg, msg)
				}
			} else if tt.expectedErrors > 0 {
				t.Errorf("expected error for key %q to exist", tt.key)
			}
		})
	}
}

func TestValidator_Check(t *testing.T) {
	tests := []struct {
		name           string
		condition      bool
		key            string
		message        string
		expectedErrors int
		expectedMsg    string
	}{
		{
			name:           "condition true - no error added",
			condition:      true,
			key:            "field",
			message:        "error message",
			expectedErrors: 0,
		},
		{
			name:           "condition false - error added",
			condition:      false,
			key:            "field",
			message:        "error message",
			expectedErrors: 1,
			expectedMsg:    "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			v.Check(tt.condition, tt.key, tt.message)

			if len(v.Errors) != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d", tt.expectedErrors, len(v.Errors))
			}

			if tt.expectedErrors > 0 {
				if msg, exists := v.Errors[tt.key]; !exists {
					t.Errorf("expected error for key %q to exist", tt.key)
				} else if msg != tt.expectedMsg {
					t.Errorf("expected message %q, got %q", tt.expectedMsg, msg)
				}
			}
		})
	}
}

func TestPermittedValue(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		permitted      []string
		expectedResult bool
	}{
		{
			name:           "value in permitted list",
			value:          "a",
			permitted:      []string{"a", "b", "c"},
			expectedResult: true,
		},
		{
			name:           "value not in permitted list",
			value:          "d",
			permitted:      []string{"a", "b", "c"},
			expectedResult: false,
		},
		{
			name:           "empty permitted list",
			value:          "a",
			permitted:      []string{},
			expectedResult: false,
		},
		{
			name:           "single permitted value - match",
			value:          "a",
			permitted:      []string{"a"},
			expectedResult: true,
		},
		{
			name:           "single permitted value - no match",
			value:          "b",
			permitted:      []string{"a"},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.PermittedValue(tt.value, tt.permitted...)
			if result != tt.expectedResult {
				t.Errorf("expected PermittedValue() to return %v, got %v", tt.expectedResult, result)
			}
		})
	}

	// Test with different types
	t.Run("integer values", func(t *testing.T) {
		result := validator.PermittedValue(5, 1, 2, 3, 4, 5)
		if !result {
			t.Error("expected PermittedValue(5, 1, 2, 3, 4, 5) to return true")
		}

		result = validator.PermittedValue(6, 1, 2, 3, 4, 5)
		if result {
			t.Error("expected PermittedValue(6, 1, 2, 3, 4, 5) to return false")
		}
	})
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name           string
		values         []string
		expectedResult bool
	}{
		{
			name:           "unique values",
			values:         []string{"a", "b", "c"},
			expectedResult: true,
		},
		{
			name:           "duplicate values",
			values:         []string{"a", "b", "a"},
			expectedResult: false,
		},
		{
			name:           "empty slice",
			values:         []string{},
			expectedResult: true,
		},
		{
			name:           "single value",
			values:         []string{"a"},
			expectedResult: true,
		},
		{
			name:           "all duplicates",
			values:         []string{"a", "a", "a"},
			expectedResult: false,
		},
		{
			name:           "two duplicates",
			values:         []string{"a", "a"},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Unique(tt.values)
			if result != tt.expectedResult {
				t.Errorf("expected Unique() to return %v, got %v", tt.expectedResult, result)
			}
		})
	}

	t.Run("integer values", func(t *testing.T) {
		result := validator.Unique([]int{1, 2, 3})
		if !result {
			t.Error("expected Unique([1, 2, 3]) to return true")
		}

		result = validator.Unique([]int{1, 2, 1})
		if result {
			t.Error("expected Unique([1, 2, 1]) to return false")
		}
	})

	t.Run("empty integer slice", func(t *testing.T) {
		result := validator.Unique([]int{})
		if !result {
			t.Error("expected Unique([]) to return true")
		}
	})
}
