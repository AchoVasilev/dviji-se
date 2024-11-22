package common

import (
	"net/http"
	"server/common"

	"github.com/gin-gonic/gin"
)

type ApiResponse map[string]any

type JSONSuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type JSONFailedValidationResponse struct {
	Success bool                     `json:"success"`
	Errors  []common.ValidationError `json:"errors"`
}

type JSONErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func SendSuccessResponse(ctx *gin.Context, message string, data interface{}, statusCode int) {
	ctx.JSON(statusCode, JSONSuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SendFailedValidationResponse(ctx *gin.Context, errors []common.ValidationError) {
	ctx.JSON(http.StatusUnprocessableEntity, JSONFailedValidationResponse{
		Success: false,
		Errors:  errors,
	})
}

func SendErrorResponse(ctx *gin.Context, message string, statusCode int) {
	ctx.JSON(statusCode, JSONErrorResponse{
		Success: false,
		Message: message,
	})
}

func SendNotFoundResponse(ctx *gin.Context, message string) {
	SendErrorResponse(ctx, message, http.StatusNotFound)
}

func SendBadRequestResponse(ctx *gin.Context, message string) {
	SendErrorResponse(ctx, message, http.StatusBadRequest)
}

func SendInternalServerResponse(ctx *gin.Context) {
	SendErrorResponse(ctx, "internal.server.error", http.StatusInternalServerError)
}
