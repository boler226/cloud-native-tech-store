package httpx

import (
	"context"
	"errors"

	"tech-store-go/internal/model"
)

type ctxKey int

const (
	idKey ctxKey = iota
	userInputKey
)

// Валідовані дані передаються від middleware до контролера через context запиту.

func WithID(ctx context.Context, id int) context.Context { return context.WithValue(ctx, idKey, id) }

func WithUserInput(ctx context.Context, in model.UserInput) context.Context {
	return context.WithValue(ctx, userInputKey, in)
}

var errMissingValidation = errors.New("validated data is missing from request context (is the validation middleware attached?)")

// ID повертає перевірений id із контексту.
func ID(ctx context.Context) (int, error) {
	id, ok := ctx.Value(idKey).(int)
	if !ok {
		return 0, errMissingValidation
	}
	return id, nil
}

// UserInput повертає перевірене тіло запиту із контексту.
func UserInput(ctx context.Context) (model.UserInput, error) {
	in, ok := ctx.Value(userInputKey).(model.UserInput)
	if !ok {
		return model.UserInput{}, errMissingValidation
	}
	return in, nil
}
