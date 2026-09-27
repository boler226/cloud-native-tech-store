// Package httpx — допоміжні функції для JSON-відповідей.
package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorBody struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

// WriteJSON серіалізує v у JSON і відправляє з потрібним статусом.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false) // не перетворювати >, < на \u003e, \u003c
	if err := enc.Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

// WriteError відправляє помилку у єдиному форматі {"error": "...", "details": [...]}.
func WriteError(w http.ResponseWriter, status int, message string, details ...string) {
	WriteJSON(w, status, errorBody{Error: message, Details: details})
}
