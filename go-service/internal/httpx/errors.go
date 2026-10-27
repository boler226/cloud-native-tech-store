package httpx

import (
	"errors"
	"log"
	"net/http"

	"tech-store-go/internal/model"
)

// ErrorBody — єдиний формат відповіді з помилкою для всього API.
type ErrorBody struct {
	Error   string `json:"error"`             // машинозчитуваний код, напр. "user_not_found"
	Message string `json:"message"`           // повідомлення для людини
	Details any    `json:"details,omitempty"` // напр. помилки полів
}

// APIError — помилка HTTP-рівня (зламаний JSON, некоректний id тощо), яку формують middleware/контролери.
type APIError struct {
	Status  int
	Code    string
	Message string
	Details any
}

func (e *APIError) Error() string { return e.Message }

func NewAPIError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

// CodeForStatus повертає код помилки за замовчуванням для HTTP-статусу.
func CodeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "method_not_allowed"
	case http.StatusConflict:
		return "conflict"
	case http.StatusRequestEntityTooLarge:
		return "payload_too_large"
	case http.StatusUnsupportedMediaType:
		return "unsupported_media_type"
	}
	if status >= 500 {
		return "internal_error"
	}
	return "bad_request"
}

// ErrorResponse — ЄДИНЕ місце, де помилки перетворюються на HTTP-статус і тіло відповіді.
// Чиста функція: легко тестується без HTTP-сервера.
func ErrorResponse(err error) (int, ErrorBody) {
	var (
		apiErr     *APIError
		validation *model.ValidationError
		notFound   *model.UserNotFoundError
		emailTaken *model.EmailTakenError
		tooLarge   *http.MaxBytesError
	)

	switch {
	case errors.As(err, &apiErr):
		return apiErr.Status, ErrorBody{Error: apiErr.Code, Message: apiErr.Message, Details: apiErr.Details}

	case errors.As(err, &validation):
		code := validation.Code
		if code == "" {
			code = "validation_failed"
		}
		body := ErrorBody{Error: code, Message: validation.Error()}
		if len(validation.Fields) > 0 {
			body.Details = validation.Fields
		}
		return http.StatusBadRequest, body

	case errors.As(err, &notFound):
		return http.StatusNotFound, ErrorBody{Error: "user_not_found", Message: notFound.Error()}

	case errors.As(err, &emailTaken):
		return http.StatusConflict, ErrorBody{Error: "email_already_exists", Message: emailTaken.Error()}

	case errors.As(err, &tooLarge):
		return http.StatusRequestEntityTooLarge, ErrorBody{Error: "payload_too_large", Message: "Request body is too large"}
	}

	// Непередбачена помилка: деталі лише в лог, клієнту — загальне повідомлення.
	return http.StatusInternalServerError, ErrorBody{Error: "internal_error", Message: "Internal server error"}
}

// WriteAppError відправляє відповідь із помилкою; 5xx додатково логує.
func WriteAppError(w http.ResponseWriter, err error) {
	status, body := ErrorResponse(err)
	if status >= http.StatusInternalServerError {
		log.Printf("internal error: %v", err)
	}
	WriteJSON(w, status, body)
}

// HandlerFunc — обробник, що ПОВЕРТАЄ помилку (замість того, щоб самому її форматувати).
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Wrap перетворює HandlerFunc на http.HandlerFunc: будь-яка повернута помилка автоматично
// проходить через ErrorResponse. Так уникається дублювання логіки помилок у кожному endpoint.
func Wrap(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			WriteAppError(w, err)
		}
	}
}
