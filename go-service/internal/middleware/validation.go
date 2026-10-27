package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strconv"

	"tech-store-go/internal/httpx"
	"tech-store-go/internal/model"
)

const maxBodyBytes = 1 << 20 // 1 MB

var idPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

// ValidateID — Middleware №2 (частина 1): перевіряє {id} у шляху і кладе число в context.
// Некоректний id -> 400 invalid_id, контролер і сервіс НЕ викликаються.
func ValidateID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.PathValue("id")
		id, err := strconv.Atoi(raw)
		if err != nil || !idPattern.MatchString(raw) {
			httpx.WriteAppError(w, httpx.NewAPIError(http.StatusBadRequest, "invalid_id",
				"Invalid identifier: must be a positive integer"))
			return
		}
		next.ServeHTTP(w, r.WithContext(httpx.WithID(r.Context(), id)))
	})
}

// ValidateUserBody — Middleware №2 (частина 2): перевіряє тіло POST/PUT:
// Content-Type, що це коректний JSON-об'єкт, обов'язкові поля, типи, порожні значення.
// Результат кладе в context; некоректний запит далі не потрапляє.
func ValidateUserBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in, err := parseUserBody(w, r)
		if err != nil {
			httpx.WriteAppError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(httpx.WithUserInput(r.Context(), in)))
	})
}

func parseUserBody(w http.ResponseWriter, r *http.Request) (model.UserInput, error) {
	bodyRequired := &model.ValidationError{Fields: []model.FieldError{{Field: "body", Message: "request body must be a JSON object"}}}

	// Тіла немає взагалі — це помилка валідації, а не "не той Content-Type".
	if r.Body == nil || r.Body == http.NoBody {
		return model.UserInput{}, bodyRequired
	}
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		return model.UserInput{}, httpx.NewAPIError(http.StatusUnsupportedMediaType, "unsupported_media_type",
			"Content-Type must be application/json")
	}

	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		return model.UserInput{}, err // *http.MaxBytesError -> 413, інакше 500
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return model.UserInput{}, bodyRequired
	}
	// Як і express.json() у Node: приймаємо лише об'єкт або масив на верхньому рівні.
	invalidJSON := httpx.NewAPIError(http.StatusBadRequest, "invalid_json", "Request body contains invalid JSON")
	if data[0] != '{' && data[0] != '[' {
		return model.UserInput{}, invalidJSON
	}

	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return model.UserInput{}, invalidJSON
	}
	obj, ok := raw.(map[string]any)
	if !ok { // масив
		return model.UserInput{}, bodyRequired
	}
	return model.ParseUserInput(obj)
}
