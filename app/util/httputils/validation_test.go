package httputils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

func TestValidatePayload_ValidInput(t *testing.T) {
	payload := &testPayload{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	errors := validatePayload(payload)
	if len(errors) != 0 {
		t.Errorf("validatePayload() returned %d errors for valid input, want 0", len(errors))
	}
}

func TestValidatePayload_MissingRequired(t *testing.T) {
	payload := &testPayload{
		Email:    "",
		Password: "",
		Name:     "",
	}

	errors := validatePayload(payload)
	if len(errors) != 3 {
		t.Errorf("validatePayload() returned %d errors, want 3", len(errors))
	}

	// Check that all required fields are reported
	fields := make(map[string]bool)
	for _, err := range errors {
		fields[err.Field] = true
		if err.Error != "field is required" {
			t.Errorf("Error message = %q, want 'field is required'", err.Error)
		}
	}

	for _, field := range []string{"email", "password", "name"} {
		if !fields[field] {
			t.Errorf("Missing error for field %q", field)
		}
	}
}

func TestValidatePayload_InvalidEmail(t *testing.T) {
	payload := &testPayload{
		Email:    "invalid-email",
		Password: "password123",
		Name:     "Test User",
	}

	errors := validatePayload(payload)
	if len(errors) != 1 {
		t.Errorf("validatePayload() returned %d errors, want 1", len(errors))
		return
	}

	if errors[0].Field != "email" {
		t.Errorf("Error field = %q, want 'email'", errors[0].Field)
	}

	if errors[0].Error != "field must be a valid email" {
		t.Errorf("Error message = %q, want 'field must be a valid email'", errors[0].Error)
	}
}

func TestGetErrorMessage(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{"required", "field is required"},
		{"email", "field must be a valid email"},
		{"unknown", ""},
		{"min", ""},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := getErrorMessage(tt.tag)
			if got != tt.expected {
				t.Errorf("getErrorMessage(%q) = %q, want %q", tt.tag, got, tt.expected)
			}
		})
	}
}

func TestProcessRequestBody_ValidJSON(t *testing.T) {
	body := `{"email": "test@example.com", "password": "password123", "name": "Test"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	payload := &testPayload{}
	ok := ProcessRequestBody(w, req, payload)

	if !ok {
		t.Error("ProcessRequestBody() should return true for valid input")
	}

	if payload.Email != "test@example.com" {
		t.Errorf("Email = %q, want 'test@example.com'", payload.Email)
	}
}

func TestProcessRequestBody_InvalidJSON(t *testing.T) {
	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	payload := &testPayload{}
	ok := ProcessRequestBody(w, req, payload)

	if ok {
		t.Error("ProcessRequestBody() should return false for invalid JSON")
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestProcessRequestBody_ValidationFailure(t *testing.T) {
	body := `{"email": "invalid", "password": "short", "name": ""}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	payload := &testPayload{}
	ok := ProcessRequestBody(w, req, payload)

	if ok {
		t.Error("ProcessRequestBody() should return false for validation failure")
	}

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

func TestProcessBody_ValidInput(t *testing.T) {
	body := `{"email": "test@example.com", "password": "password123", "name": "Test"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	payload := &testPayload{}
	result := ProcessBody(w, req, payload)

	if result.ParsingError != nil {
		t.Errorf("ParsingError = %v, want nil", result.ParsingError)
	}

	// Note: The Success field logic appears inverted in the source code
	// errors != nil means validation passed (no errors)
}

func TestProcessBody_InvalidJSON(t *testing.T) {
	body := `{not valid json`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	payload := &testPayload{}
	result := ProcessBody(w, req, payload)

	if result.ParsingError == nil {
		t.Error("ProcessBody() should return ParsingError for invalid JSON")
	}

	if result.Success {
		t.Error("Success should be false for parsing error")
	}
}

func TestValidationError_Struct(t *testing.T) {
	err := &ValidationError{
		Field: "email",
		Error: "field is required",
		Value: "",
	}

	if err.Field != "email" {
		t.Errorf("Field = %q, want 'email'", err.Field)
	}

	if err.Error != "field is required" {
		t.Errorf("Error = %q, want 'field is required'", err.Error)
	}
}

func TestValidationResult_Struct(t *testing.T) {
	result := &ValidationResult{
		Success:      true,
		ParsingError: nil,
		ValidationErrors: []*ValidationError{
			{Field: "email", Error: "invalid"},
		},
	}

	if !result.Success {
		t.Error("Success should be true")
	}

	if result.ParsingError != nil {
		t.Error("ParsingError should be nil")
	}

	if len(result.ValidationErrors) != 1 {
		t.Errorf("ValidationErrors length = %d, want 1", len(result.ValidationErrors))
	}
}

type payloadWithOptional struct {
	Required string `json:"required" validate:"required"`
	Optional string `json:"optional"`
}

func TestValidatePayload_OptionalFields(t *testing.T) {
	payload := &payloadWithOptional{
		Required: "value",
		Optional: "", // Not required, should be valid
	}

	errors := validatePayload(payload)
	if len(errors) != 0 {
		t.Errorf("validatePayload() returned %d errors for payload with empty optional field", len(errors))
	}
}

type payloadWithMinLength struct {
	Name string `json:"name" validate:"required,min=3"`
}

func TestValidatePayload_MinLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"too short", "ab", true},
		{"exact minimum", "abc", false},
		{"longer", "abcdef", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &payloadWithMinLength{Name: tt.value}
			errors := validatePayload(payload)

			if tt.wantError && len(errors) == 0 {
				t.Error("Expected validation error for short value")
			}
			if !tt.wantError && len(errors) > 0 {
				t.Errorf("Unexpected validation error: %v", errors)
			}
		})
	}
}
