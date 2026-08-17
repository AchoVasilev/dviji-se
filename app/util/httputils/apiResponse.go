package httputils

import (
	"log/slog"
	"net/http"
	"server/util/ctxutils"
	"server/util/jsonutils"
)

type (
	ApiResponse         map[string]any
	JSONSuccessResponse struct {
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
)

type JSONFailedValidationResponse struct {
	Success bool               `json:"success"`
	Errors  []*ValidationError `json:"errors"`
}

type JSONErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func SendApiResponse(writer http.ResponseWriter, req *http.Request, status int, body any, message string) {
	switch status {
	case http.StatusOK:
		SendOkWithBody(writer, body)
	case http.StatusCreated:
		SendCreatedAt(writer, message)
	case http.StatusNotFound:
		SendNotFoundResponse(writer, message)
	case http.StatusInternalServerError:
		SendInternalServerResponse(writer, req)
	case http.StatusConflict:
		SendConflictResponse(writer, message)
	case http.StatusBadRequest:
		SendBadRequestResponse(writer, message)
	default:
		SendOkWithBody(writer, nil)
	}
}

func SendCreatedAt(writer http.ResponseWriter, uri string) {
	err := jsonutils.WriteCreatedAt(writer, uri, nil)
	if err != nil {
		slog.Warn("Failed to write response body", "error", err)
	}
}

func SendOkWithBody(writer http.ResponseWriter, data interface{}) {
	err := jsonutils.WriteJSON(writer, http.StatusOK, data)
	if err != nil {
		slog.Warn("Failed to write response body", "error", err)
	}
}

func SendSuccessResponse(writer http.ResponseWriter, message string, data interface{}, statusCode int) {
	err := jsonutils.WriteJSON(writer, statusCode, JSONSuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
	if err != nil {
		slog.Warn("Failed to write response body", "error", err)
	}
}

func SendFailedValidationResponse(writer http.ResponseWriter, errors []*ValidationError) {
	err := jsonutils.WriteJSON(writer, http.StatusUnprocessableEntity, JSONFailedValidationResponse{
		Success: false,
		Errors:  errors,
	})
	if err != nil {
		slog.Warn("Failed to write response body", "error", err)
	}
}

func SendErrorResponse(writer http.ResponseWriter, message string, statusCode int) {
	err := jsonutils.WriteJSON(writer, statusCode, JSONErrorResponse{
		Success: false,
		Message: message,
	})
	if err != nil {
		slog.Warn("Failed to write response body", "error", err)
	}
}

func SendNotFoundResponse(writer http.ResponseWriter, message string) {
	SendErrorResponse(writer, message, http.StatusNotFound)
}

func SendBadRequestResponse(writer http.ResponseWriter, message string) {
	SendErrorResponse(writer, message, http.StatusBadRequest)
}

func SendInternalServerResponse(writer http.ResponseWriter, req *http.Request) {
	reqId := ctxutils.RequestIdFromContext(req.Context())
	writer.Header().Set("X-REQUEST-ID", reqId)
	SendErrorResponse(writer, "internal.server.error", http.StatusInternalServerError)
}

func SendConflictResponse(writer http.ResponseWriter, message string) {
	SendErrorResponse(writer, message, http.StatusConflict)
}
