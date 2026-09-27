// Package service — бізнес-логіка: валідація та правила роботи з товарами.
package service

import (
	"errors"
	"strings"
	"unicode/utf8"

	"tech-store-go/internal/model"
	"tech-store-go/internal/repository"
)

// ErrNotFound — товар не знайдено (handler перетворює на 404).
var ErrNotFound = errors.New("product not found")

// ValidationError — помилки валідації (handler перетворює на 400).
type ValidationError struct {
	Details []string
}

func (e *ValidationError) Error() string { return "validation failed" }

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(category string) []model.Product {
	return s.repo.FindAll(category)
}

func (s *ProductService) Get(id int64) (model.Product, error) {
	p, ok := s.repo.FindByID(id)
	if !ok {
		return model.Product{}, ErrNotFound
	}
	return p, nil
}

func (s *ProductService) Create(in model.ProductInput) (model.Product, error) {
	p, err := validate(in)
	if err != nil {
		return model.Product{}, err
	}
	return s.repo.Create(p), nil
}

func (s *ProductService) Update(id int64, in model.ProductInput) (model.Product, error) {
	p, err := validate(in)
	if err != nil {
		return model.Product{}, err
	}
	updated, ok := s.repo.Update(id, p)
	if !ok {
		return model.Product{}, ErrNotFound
	}
	return updated, nil
}

func (s *ProductService) Delete(id int64) error {
	if !s.repo.Delete(id) {
		return ErrNotFound
	}
	return nil
}

// validate перевіряє вхідні дані й повертає готовий до збереження Product.
// Правила ті самі, що й у Node-версії.
func validate(in model.ProductInput) (model.Product, error) {
	var errs []string

	name := trimmed(in.Name)
	if name == "" || utf8.RuneCountInString(name) > 100 {
		errs = append(errs, "name is required and must be a non-empty string (max 100 chars)")
	}
	category := trimmed(in.Category)
	if category == "" {
		errs = append(errs, "category is required and must be a non-empty string")
	}
	brand := trimmed(in.Brand)
	if brand == "" {
		errs = append(errs, "brand is required and must be a non-empty string")
	}
	if in.Price == nil || *in.Price <= 0 {
		errs = append(errs, "price is required and must be a number greater than 0")
	}
	if in.Stock == nil || *in.Stock < 0 {
		errs = append(errs, "stock is required and must be an integer >= 0")
	}

	if len(errs) > 0 {
		return model.Product{}, &ValidationError{Details: errs}
	}

	return model.Product{
		Name:        name,
		Category:    category,
		Brand:       brand,
		Price:       *in.Price,
		Stock:       *in.Stock,
		Description: trimmed(in.Description),
	}, nil
}

func trimmed(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}
