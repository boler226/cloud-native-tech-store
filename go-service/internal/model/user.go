package model

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	NameMaxLen  = 100
	EmailMaxLen = 254
)

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// User — сутність "користувач" (клієнт магазину).
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserInput — вхідні дані для створення/оновлення користувача.
type UserInput struct {
	Name  string
	Email string
}

// NormalizeUserInput — бізнес-правило: name без пробілів по краях, email у нижньому регістрі.
func NormalizeUserInput(in UserInput) UserInput {
	return UserInput{
		Name:  strings.TrimSpace(in.Name),
		Email: strings.ToLower(strings.TrimSpace(in.Email)),
	}
}

// ParseUserInput перевіряє "сирий" JSON-об'єкт: обов'язковість полів, типи, порожні значення,
// довжину та формат. Збирає ВСІ помилки одразу, а не лише першу.
func ParseUserInput(raw map[string]any) (UserInput, error) {
	if raw == nil {
		return UserInput{}, &ValidationError{Fields: []FieldError{{Field: "body", Message: "request body must be a JSON object"}}}
	}

	var errs []FieldError
	name, nameOK := readString(raw, "name", &errs)
	email, emailOK := readString(raw, "email", &errs)
	if nameOK {
		errs = append(errs, checkName(name)...)
	}
	if emailOK {
		errs = append(errs, checkEmail(email)...)
	}
	if len(errs) > 0 {
		return UserInput{}, &ValidationError{Fields: errs}
	}
	return UserInput{Name: name, Email: email}, nil
}

// ValidateUserInput перевіряє вже типізовані дані (використовує Service, щоб захищати інваріанти
// й тоді, коли його викликають не через HTTP).
func ValidateUserInput(in UserInput) error {
	errs := append(checkName(in.Name), checkEmail(in.Email)...)
	if len(errs) > 0 {
		return &ValidationError{Fields: errs}
	}
	return nil
}

func readString(raw map[string]any, field string, errs *[]FieldError) (string, bool) {
	v, present := raw[field]
	if !present || v == nil {
		*errs = append(*errs, FieldError{Field: field, Message: "is required"})
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		*errs = append(*errs, FieldError{Field: field, Message: "must be a string"})
		return "", false
	}
	return s, true
}

func checkName(s string) []FieldError {
	t := strings.TrimSpace(s)
	switch {
	case t == "":
		return []FieldError{{Field: "name", Message: "must not be empty"}}
	case utf8.RuneCountInString(t) > NameMaxLen:
		return []FieldError{{Field: "name", Message: "must be at most 100 characters"}}
	}
	return nil
}

func checkEmail(s string) []FieldError {
	t := strings.TrimSpace(s)
	switch {
	case t == "":
		return []FieldError{{Field: "email", Message: "must not be empty"}}
	case utf8.RuneCountInString(t) > EmailMaxLen:
		return []FieldError{{Field: "email", Message: "must be at most 254 characters"}}
	case !emailRe.MatchString(t):
		return []FieldError{{Field: "email", Message: "must be a valid email address"}}
	}
	return nil
}
