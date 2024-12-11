package middleware

import (
	"net/http"
	"server/common"
	"server/common/api"
	"server/infrastructure/utils"
)

func ValidateBody[T any](next func(writer http.ResponseWriter, req *http.Request, _ *T)) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		params := new(T)
		if err := utils.ParseJSON(req, params); err != nil {
			api.SendInternalServerResponse(writer)
			return
		}

		if err := common.ValidateRequestBody(params); err != nil {
			api.SendFailedValidationResponse(writer, err)
			return
		}

		next(writer, req, params)
	}
}
