package httputils

import (
	"context"
	"net/http"
	"server/util/jsonutils"
)

type ApiResponse map[string]any
type JSONSuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type JSONFailedValidationResponse struct {
	Success bool               `json:"success"`
	Errors  []*ValidationError `json:"errors"`
}

type JSONErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func SendOkWithBody(writer http.ResponseWriter, data interface{}) {
	jsonutils.WriteJSON(writer, http.StatusOK, data)
}

func SendSuccessResponse(writer http.ResponseWriter, message string, data interface{}, statusCode int) {
	jsonutils.WriteJSON(writer, statusCode, JSONSuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SendFailedValidationResponse(writer http.ResponseWriter, errors []*ValidationError) {
	jsonutils.WriteJSON(writer, http.StatusUnprocessableEntity, JSONFailedValidationResponse{
		Success: false,
		Errors:  errors,
	})
}

func SendErrorResponse(writer http.ResponseWriter, message string, statusCode int) {
	jsonutils.WriteJSON(writer, statusCode, JSONErrorResponse{
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

func SendInternalServerResponse(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	reqId := requestIdFromContext(ctx)
	writer.Header().Set("X-REQUEST-ID", reqId)
	SendErrorResponse(writer, "internal.server.error", http.StatusInternalServerError)
}

func requestIdFromContext(ctx context.Context) string {
	value := ctx.Value("requestId")
	if value == nil {
		return ""
	}

	id, ok := value.(string)
	if !ok {
		return ""
	}

	return id
}
