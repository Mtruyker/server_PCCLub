package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Standard API response wrapper for mobile app compatibility
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Code    string      `json:"code,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// writeAPISuccess writes a successful API response with data wrapper
func writeAPISuccess(w http.ResponseWriter, status int, data interface{}) {
	response := APIResponse{
		Data:    data,
		Success: true,
	}
	writeJSON(w, status, response)
}

// writeAPIResult writes a successful API response with result wrapper
func writeAPIResult(w http.ResponseWriter, status int, result interface{}) {
	response := APIResponse{
		Result:  result,
		Success: true,
	}
	writeJSON(w, status, response)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeErrorCode(w, status, "ERROR", message)
}

func writeErrorCode(w http.ResponseWriter, status int, code string, message string) {
	response := APIResponse{
		Success: false,
		Message: message,
		Code:    code,
	}
	writeJSON(w, status, response)
}

// writeAPIError writes a standardized error response
func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeErrorCode(w, status, code, message)
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func pathID(r *http.Request) (int64, error) {
	raw := pathValue(r, "id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func requireText(value string, field string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(field + " is required")
	}
	return nil
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
