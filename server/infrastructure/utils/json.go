package utils

import (
	"encoding/json"
	"fmt"
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
		return fmt.Errorf("Request body is requred")
	}

	return json.NewDecoder(req.Body).Decode(&value)
}
