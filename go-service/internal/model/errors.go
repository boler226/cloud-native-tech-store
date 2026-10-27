package model

import (
	"errors"
	"fmt"
)

// Узагальнені sentinel-помилки для перевірок через errors.Is (без прив'язки до конкретного id/email).
var (
	ErrNotFound   = errors.New("not found")
	ErrEmailTaken = errors.New("email already in use")
)

// UserNotFoundError — користувача з таким id не існує (HTTP-шар перетворить на 404).
type UserNotFoundError struct{ ID int }

func (e *UserNotFoundError) Error() string {
	if e.ID == 0 { // пошук за іншою ознакою (напр. email) — id невідомий
		return "User was not found"
	}
	return fmt.Sprintf("User with id %d was not found", e.ID)
}

func (e *UserNotFoundError) Is(target error) bool { return target == ErrNotFound }

// EmailTakenError — email уже використовується іншим користувачем (-> 409).
type EmailTakenError struct{ Email string }

func (e *EmailTakenError) Error() string {
	return fmt.Sprintf("User with email %s already exists", e.Email)
}

func (e *EmailTakenError) Is(target error) bool { return target == ErrEmailTaken }

// FieldError — помилка конкретного поля вхідних даних.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError — некоректні вхідні дані (-> 400).
// Code і Message необов'язкові: за замовчуванням "validation_failed" / "Request validation failed".
type ValidationError struct {
	Code    string
	Message string
	Fields  []FieldError
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Request validation failed"
}
