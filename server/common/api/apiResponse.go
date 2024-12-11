package api

import (
	"net/http"
	"server/common"
	"server/infrastructure/utils"
)

type ApiResponse map[string]any

type JSONSuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type JSONFailedValidationResponse struct {
	Success bool                      `json:"success"`
	Errors  []*common.ValidationError `json:"errors"`
}

type JSONErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func SendOkWithBody(writer http.ResponseWriter, data interface{}) {
	utils.WriteJSON(writer, http.StatusOK, data)
}

func SendSuccessResponse(writer http.ResponseWriter, message string, data interface{}, statusCode int) {
	utils.WriteJSON(writer, statusCode, JSONSuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SendFailedValidationResponse(writer http.ResponseWriter, errors []*common.ValidationError) {
	utils.WriteJSON(writer, http.StatusUnprocessableEntity, JSONFailedValidationResponse{
		Success: false,
		Errors:  errors,
	})
}

func SendErrorResponse(writer http.ResponseWriter, message string, statusCode int) {
	utils.WriteJSON(writer, statusCode, JSONErrorResponse{
		Success: false,
		Message: message,
	})
}

func SendNotFoundResponse(writer http.ResponseWriter, message string) {
	SendErrorResponse(writer, message, http.StatusNotFound)
}

func SendBadRequestResponse(writer http.ResponseWriter, message string) {
	SendErrorResponse(writer, message, http.StatusBadRequest)
}

func SendInternalServerResponse(writer http.ResponseWriter) {
	SendErrorResponse(writer, "internal.server.error", http.StatusInternalServerError)
}
