// Package httpx — допоміжні функції для JSON-відповідей.
package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

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

// WriteError відправляє помилку у єдиному форматі {"error": "<код>", "message": "...", "details": [...]}.
// Код визначається за статусом (400 з details -> "validation_failed"). Використовується
// обробниками, що не повертають error (products, concurrency, router fallback-и).
func WriteError(w http.ResponseWriter, status int, message string, details ...string) {
	body := ErrorBody{Error: CodeForStatus(status), Message: message}
	if len(details) > 0 {
		body.Details = details
		if status == http.StatusBadRequest {
			body.Error = "validation_failed"
		}
	}
	WriteJSON(w, status, body)
}
