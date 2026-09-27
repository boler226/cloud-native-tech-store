// Package repository — шар доступу до даних (тут: in-memory сховище).
package repository

import (
	"sort"
	"strings"
	"sync"
	"time"

	"tech-store-go/internal/model"
)

// ProductRepository — інтерфейс сховища. Завдяки йому in-memory реалізацію
// можна замінити на БД, не чіпаючи сервіс.
type ProductRepository interface {
	FindAll(category string) []model.Product
	FindByID(id int64) (model.Product, bool)
	Create(p model.Product) model.Product
	Update(id int64, p model.Product) (model.Product, bool)
	Delete(id int64) bool
}

// MemoryProductRepository зберігає товари в map. Mutex потрібен, бо
// net/http обробляє кожен запит в окремій горутині.
type MemoryProductRepository struct {
	mu     sync.RWMutex
	items  map[int64]model.Product
	nextID int64
}

func NewMemoryProductRepository(seed []model.Product) *MemoryProductRepository {
	r := &MemoryProductRepository{items: make(map[int64]model.Product), nextID: 1}
	for _, p := range seed {
		r.Create(p)
	}
	return r
}

func now() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

func (r *MemoryProductRepository) FindAll(category string) []model.Product {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]model.Product, 0, len(r.items))
	for _, p := range r.items {
		if category == "" || strings.EqualFold(p.Category, category) {
			result = append(result, p)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (r *MemoryProductRepository) FindByID(id int64) (model.Product, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.items[id]
	return p, ok
}

func (r *MemoryProductRepository) Create(p model.Product) model.Product {
	r.mu.Lock()
	defer r.mu.Unlock()

	p.ID = r.nextID
	r.nextID++
	p.CreatedAt = now()
	p.UpdatedAt = p.CreatedAt
	r.items[p.ID] = p
	return p
}

func (r *MemoryProductRepository) Update(id int64, p model.Product) (model.Product, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.items[id]
	if !ok {
		return model.Product{}, false
	}
	p.ID = existing.ID
	p.CreatedAt = existing.CreatedAt
	p.UpdatedAt = now()
	r.items[id] = p
	return p, true
}

func (r *MemoryProductRepository) Delete(id int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return false
	}
	delete(r.items, id)
	return true
}
