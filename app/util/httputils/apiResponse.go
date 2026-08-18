package httputils

import (
	"context"
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

func SendApiResponse(ctx context.Context, writer http.ResponseWriter, req *http.Request, status int, body any, message string) {
	switch status {
	case http.StatusOK:
		SendOkWithBody(ctx, writer, body)
	case http.StatusCreated:
		SendCreatedAt(ctx, writer, message)
	case http.StatusNotFound:
		SendNotFoundResponse(ctx, writer, message)
	case http.StatusInternalServerError:
		SendInternalServerResponse(writer, req)
	case http.StatusConflict:
		SendConflictResponse(ctx, writer, message)
	case http.StatusBadRequest:
		SendBadRequestResponse(ctx, writer, message)
	default:
		SendOkWithBody(ctx, writer, nil)
	}
}

func SendCreatedAt(ctx context.Context, writer http.ResponseWriter, uri string) {
	err := jsonutils.WriteCreatedAt(writer, uri, nil)
	if err != nil {
		slog.WarnContext(ctx, "Failed to write response body", "error", err)
	}
}

func SendOkWithBody(ctx context.Context, writer http.ResponseWriter, data interface{}) {
	err := jsonutils.WriteJSON(writer, http.StatusOK, data)
	if err != nil {
		slog.WarnContext(ctx, "Failed to write response body", "error", err)
	}
}

func SendSuccessResponse(ctx context.Context, writer http.ResponseWriter, message string, data interface{}, statusCode int) {
	err := jsonutils.WriteJSON(writer, statusCode, JSONSuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
	if err != nil {
		slog.WarnContext(ctx, "Failed to write response body", "error", err)
	}
}

func SendFailedValidationResponse(ctx context.Context, writer http.ResponseWriter, errors []*ValidationError) {
	err := jsonutils.WriteJSON(writer, http.StatusUnprocessableEntity, JSONFailedValidationResponse{
		Success: false,
		Errors:  errors,
	})
	if err != nil {
		slog.WarnContext(ctx, "Failed to write response body", "error", err)
	}
}

func SendErrorResponse(ctx context.Context, writer http.ResponseWriter, message string, statusCode int) {
	err := jsonutils.WriteJSON(writer, statusCode, JSONErrorResponse{
		Success: false,
		Message: message,
	})
	if err != nil {
		slog.WarnContext(ctx, "Failed to write response body", "error", err)
	}
}

func SendNotFoundResponse(ctx context.Context, writer http.ResponseWriter, message string) {
	SendErrorResponse(ctx, writer, message, http.StatusNotFound)
}

func SendBadRequestResponse(ctx context.Context, writer http.ResponseWriter, message string) {
	SendErrorResponse(ctx, writer, message, http.StatusBadRequest)
}

func SendInternalServerResponse(writer http.ResponseWriter, req *http.Request) {
	reqId := ctxutils.RequestIdFromContext(req.Context())
	writer.Header().Set("X-REQUEST-ID", reqId)
	SendErrorResponse(req.Context(), writer, "internal.server.error", http.StatusInternalServerError)
}

func SendConflictResponse(ctx context.Context, writer http.ResponseWriter, message string) {
	SendErrorResponse(ctx, writer, message, http.StatusConflict)
}
