package seed

import "tech-store-go/internal/model"

// Users — початкові користувачі (клієнти магазину).
func Users() []model.UserInput {
	return []model.UserInput{
		{Name: "John Doe", Email: "john@example.com"},
		{Name: "Jane Smith", Email: "jane@example.com"},
	}
}
