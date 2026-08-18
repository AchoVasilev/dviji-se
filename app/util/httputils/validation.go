package httputils

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"runtime/debug"
	"server/util/jsonutils"
	"server/util/securityutil"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
	Value string `json:"value"`
}

type ValidationResult struct {
	Success          bool
	ParsingError     error
	ValidationErrors []*ValidationError
}

func ProcessRequestBody(writer http.ResponseWriter, req *http.Request, payload any) bool {
	if err := jsonutils.ParseJSON(req, payload); err != nil {
		// An oversized body is the caller's problem, not a server fault.
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			slog.WarnContext(req.Context(), "Request body too large", "limit", maxBytesErr.Limit, "path", req.URL.Path)
			SendErrorResponse(req.Context(), writer, "request.body.too.large", http.StatusRequestEntityTooLarge)
			return false
		}

		slog.ErrorContext(req.Context(), "Could not parse request body", "error", err, "stack", string(debug.Stack()))
		SendInternalServerResponse(writer, req)
		return false
	}

	if err := validatePayload(payload); err != nil {
		SendFailedValidationResponse(req.Context(), writer, err)
		return false
	}

	return true
}

func ProcessBody(writer http.ResponseWriter, req *http.Request, payload any) *ValidationResult {
	if err := jsonutils.ParseJSON(req, payload); err != nil {
		return &ValidationResult{
			Success:          false,
			ParsingError:     err,
			ValidationErrors: nil,
		}
	}

	validationErrors := validatePayload(payload)
	return &ValidationResult{
		ParsingError:     nil,
		ValidationErrors: validationErrors,
		Success:          validationErrors == nil,
	}
}

// validate is safe for concurrent use and caches struct metadata, so it is
// built once rather than per request.
var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	// "strongpassword" is for passwords being set, never for login: existing
	// accounts may hold weaker passwords and must still be able to sign in.
	if err := v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return securityutil.IsPasswordStrong(fl.Field().String())
	}); err != nil {
		panic(fmt.Sprintf("could not register the strongpassword rule: %v", err))
	}

	return v
}

func validatePayload(payload any) []*ValidationError {
	var result []*ValidationError
	err := validate.Struct(payload)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		reflected := reflect.ValueOf(payload)
		for _, validationErr := range validationErrors {
			field, _ := reflected.Elem().Type().FieldByName(validationErr.StructField())
			key := field.Tag.Get("json")
			if key == "" {
				key = strings.ToLower(validationErr.StructField())
			}

			value := ""
			// Never echo back a secret, and only report values that are already
			// strings - Value() is an any and may hold anything. Keyed on the
			// field rather than the rule, so renaming a rule cannot silently
			// start leaking the submitted password back to the caller.
			if !isSecretField(key) {
				if s, ok := validationErr.Value().(string); ok {
					value = s
				}
			}

			result = append(result, &ValidationError{
				Field: key,
				Error: getErrorMessage(validationErr),
				Value: value,
			})
		}
	}

	return result
}

// isSecretField reports whether a field's submitted value must never be echoed
// back in a validation response.
func isSecretField(field string) bool {
	lower := strings.ToLower(field)

	return strings.Contains(lower, "password") || strings.Contains(lower, "token") || strings.Contains(lower, "secret")
}

// getErrorMessage renders a human readable message for a failed rule. These
// surface in JSON responses via SendFailedValidationResponse; the HTMX form
// path renders its own copy keyed on the field name. Unknown tags must still
// produce something, otherwise the caller receives an empty string.
func getErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "field must be a valid email"
	case "min":
		return fmt.Sprintf("field must be at least %s characters", fieldErr.Param())
	case "max":
		return fmt.Sprintf("field must be at most %s characters", fieldErr.Param())
	case "oneof":
		return fmt.Sprintf("field must be one of: %s", strings.ReplaceAll(fieldErr.Param(), " ", ", "))
	case "strongpassword":
		return fmt.Sprintf(
			"password must be at least %d characters and include an upper case letter, a lower case letter, a digit and a symbol",
			securityutil.MinPasswordLength)
	case "uuid":
		return "field must be a valid UUID"
	case "url":
		return "field must be a valid URL"
	default:
		return fmt.Sprintf("field failed the %q rule", fieldErr.Tag())
	}
}
