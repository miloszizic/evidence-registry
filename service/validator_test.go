package service

import (
	"database/sql"
	"testing"
)

func TestHasErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		errors      []string
		fieldErrors map[string]string
		want        bool
	}{
		{"has no errors and field errors and returns false", nil, nil, false},
		{"has errors and returns true", []string{"error1"}, nil, true},
		{"has FieldErrors and returns true", nil, map[string]string{"field1": "error1"}, true},
		{"has both errors and returns true", []string{"error1"}, map[string]string{"field1": "error1"}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := &Validator{Errors: tt.errors, FieldErrors: tt.fieldErrors}
			if got := v.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddError(t *testing.T) {
	tests := []struct {
		name   string
		errors []string
		want   []string
	}{
		{"has one error", []string{"error1"}, []string{"error1"}},
		{"has two errors", []string{"error1", "error2"}, []string{"error1", "error2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{}
			for _, err := range tt.errors {
				v.AddError(err)
			}
			if len(v.Errors) != len(tt.want) || !equalSlice(v.Errors, tt.want) {
				t.Errorf("AddError() = %v, want %v", v.Errors, tt.want)
			}
		})
	}
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestAddFieldError(t *testing.T) {
	t.Parallel()

	v := &Validator{}

	v.AddFieldError("field1", "fieldError1")
	if len(v.FieldErrors) != 1 || v.FieldErrors["field1"] != "fieldError1" {
		t.Errorf("AddFieldError() failed, got: %v", v.FieldErrors)
	}

	v.AddFieldError("field2", "fieldError2")
	if len(v.FieldErrors) != 2 || v.FieldErrors["field2"] != "fieldError2" {
		t.Errorf("AddFieldError() failed, got: %v", v.FieldErrors)
	}
	// Try adding an error for an existing field
	v.AddFieldError("field1", "fieldError1Duplicate")
	if v.FieldErrors["field1"] != "fieldError1" {
		t.Errorf("AddFieldError() should not overwrite existing errors, got: %v", v.FieldErrors["field1"])
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()

	v := &Validator{}

	v.Check(true, "error1")
	if len(v.Errors) != 0 {
		t.Errorf("Check() should not add errors for ok=true, got: %v", v.Errors)
	}

	v.Check(false, "error1")
	if len(v.Errors) != 1 || v.Errors[0] != "error1" {
		t.Errorf("Check() failed, got: %v", v.Errors)
	}
}

func TestCheckField(t *testing.T) {
	v := &Validator{}
	v.CheckField(true, "field1", "fieldError1")
	if len(v.FieldErrors) != 0 {
		t.Errorf("CheckField() should not add errors for ok=true, got: %v", v.FieldErrors)
	}
	v.CheckField(false, "field1", "fieldError1")
	if len(v.FieldErrors) != 1 || v.FieldErrors["field1"] != "fieldError1" {
		t.Errorf("CheckField() failed, got: %v", v.FieldErrors)
	}
}

func TestCheckNullStringField(t *testing.T) {
	t.Parallel()

	v := &Validator{}
	validString := sql.NullString{String: "valid", Valid: true}
	emptyValidString := sql.NullString{String: "", Valid: true}
	invalidString := sql.NullString{Valid: false}

	v.CheckNullStringField(validString, "field1", "fieldError1")
	if len(v.FieldErrors) != 0 {
		t.Errorf("CheckNullStringField() should not add errors for valid non-empty string, got: %v", v.FieldErrors)
	}

	v.CheckNullStringField(emptyValidString, "field2", "fieldError2")
	if len(v.FieldErrors) != 1 || v.FieldErrors["field2"] != "fieldError2" {
		t.Errorf("CheckNullStringField() failed for empty valid string, got: %v", v.FieldErrors)
	}

	v.CheckNullStringField(invalidString, "field3", "fieldError3")
	if len(v.FieldErrors) != 2 || v.FieldErrors["field3"] != "fieldError3" {
		t.Errorf("CheckNullStringField() failed for invalid string, got: %v", v.FieldErrors)
	}
}
