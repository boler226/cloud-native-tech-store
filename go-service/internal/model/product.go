// Package model описує сутності предметної області.
package model

import "time"

// Product — товар інтернет-магазину техніки.
type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Brand       string    `json:"brand"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ProductInput — тіло запиту для POST/PUT. Поля-вказівники дозволяють відрізнити
// "поле не передали" (nil) від "передали нульове значення".
type ProductInput struct {
	Name        *string  `json:"name"`
	Category    *string  `json:"category"`
	Brand       *string  `json:"brand"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	Description *string  `json:"description"`
}
