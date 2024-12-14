package jsonutils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func WriteCreatedAt(writer http.ResponseWriter, uri string, value any) error {
	writer.Header().Add("Location", uri)
	return WriteJSON(writer, http.StatusCreated, value)
}

func WriteJSON(writer http.ResponseWriter, status int, value any) error {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(status)

	return json.NewEncoder(writer).Encode(value)
}

func ParseJSON(req *http.Request, value any) error {
	if req.Body == nil {
		slog.Warn("Missing request body. [requestMethod=%s, requestUri=%s]", req.Method, req.RequestURI)
		return fmt.Errorf("Body is required")
	}

	return json.NewDecoder(req.Body).Decode(&value)
}
