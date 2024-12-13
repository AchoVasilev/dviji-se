package api

import (
	"fmt"
	"net/http"
	"reflect"
	"server/infrastructure/utils"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

func ProcessRequestBody[R interface{}](writer http.ResponseWriter, req *http.Request, payload *R) bool {
	if err := utils.ParseJSON(req, payload); err != nil {
		SendInternalServerResponse(writer)
		return false
	}

	if err := validatePayload(payload); err != nil {
		SendFailedValidationResponse(writer, err)
		return false
	}

	return true
}

func validatePayload[R interface{}](payload *R) []*ValidationError {
	var validate *validator.Validate
	validate = validator.New(validator.WithRequiredStructEnabled())

	var errors []*ValidationError
	err := validate.Struct(payload)

	validationErrors, ok := err.(validator.ValidationErrors)

	if ok {
		reflected := reflect.ValueOf(payload)
		for _, validationErr := range validationErrors {
			field, _ := reflected.Type().FieldByName(validationErr.StructField())

			key := field.Tag.Get("json")
			if key == "" {
				key = strings.ToLower(validationErr.StructField())
			}

			fmt.Println(validationErr.Field())
			currentErr := &ValidationError{
				Field: key,
				Error: getErrorMessage(validationErr.Tag(), key),
			}

			errors = append(errors, currentErr)
		}
	}

	return errors
}

func getErrorMessage(tag string, field string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	}

	return ""
}
