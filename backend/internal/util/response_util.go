package util

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	CorrelationID string `json:"correlationId"`
}

func WriteSuccessJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := &APIResponse{Data: data}
	json.NewEncoder(w).Encode(response)
}

func WriteErrorJSON(w http.ResponseWriter, errorCode ErrorCode, correlationID string) {
	w.Header().Set("Content-Type", "application/json")

	error := ErrorMessages[errorCode]
	response := &APIError{
		Code:          string(error.Code),
		Message:       error.Message,
		CorrelationID: correlationID,
	}

	w.WriteHeader(error.Status)

	json.NewEncoder(w).Encode(response)
}
