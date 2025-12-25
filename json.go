package main

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// APIError represents a structured error response
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	// Get caller info for better debugging
	_, file, line, _ := runtime.Caller(1)
	// Extract just the filename
	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]

	// Log with color coding based on error severity
	if code >= 500 {
		log.Printf("%s[ERROR]%s %s:%d - %s", colorRed, colorReset, filename, line, msg)
		if err != nil {
			log.Printf("%s[CAUSE]%s %v", colorRed, colorReset, err)
		}
	} else if code >= 400 {
		log.Printf("%s[WARN]%s %s:%d - %s", colorYellow, colorReset, filename, line, msg)
		if err != nil {
			log.Printf("%s[DETAIL]%s %v", colorYellow, colorReset, err)
		}
	}

	respondWithJSON(w, code, APIError{
		Error:   http.StatusText(code),
		Message: msg,
		Code:    code,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("%s[ERROR]%s JSON marshal failed: %s", colorRed, colorReset, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error","code":500}`))
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

// respondWithSuccess is a helper for successful responses with a message
func respondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	type successResponse struct {
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}
	respondWithJSON(w, code, successResponse{
		Message: message,
		Data:    data,
	})
}
