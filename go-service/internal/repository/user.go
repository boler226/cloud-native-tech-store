package repository

import (
	"strings"
	"sync"

	"tech-store-go/internal/datastore"
	"tech-store-go/internal/model"
)

// InMemoryUserRepository — реалізація інтерфейсу service.UserRepository поверх DataStore.
//
// Цей пакет НЕ імпортує пакет service: залежність спрямована лише в один бік
// (service визначає інтерфейс, repository його реалізує). Відповідність інтерфейсу
// перевіряється компілятором там, де реалізацію передають у сервіс (app.New і тести).
type InMemoryUserRepository struct {
	mu    sync.Mutex // робить "перевірка email + запис" атомарною (аналог UNIQUE-індексу в БД)
	store *datastore.MemoryStore[model.User]
}

// NewInMemoryUserRepository створює репозиторій і (опційно) наповнює його початковими користувачами.
func NewInMemoryUserRepository(store *datastore.MemoryStore[model.User], seed ...model.UserInput) *InMemoryUserRepository {
	r := &InMemoryUserRepository{store: store}
	for _, in := range seed {
		id := store.NextID()
		store.Put(id, model.User{ID: id, Name: in.Name, Email: in.Email})
	}
	return r
}

func (r *InMemoryUserRepository) Create(user model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.emailTaken(user.Email, 0) {
		return model.User{}, &model.EmailTakenError{Email: user.Email}
	}
	user.ID = r.store.NextID()
	r.store.Put(user.ID, user)
	return user, nil
}

func (r *InMemoryUserRepository) FindByID(id int) (model.User, error) {
	u, ok := r.store.Get(id)
	if !ok {
		return model.User{}, &model.UserNotFoundError{ID: id}
	}
	return u, nil
}

func (r *InMemoryUserRepository) FindByEmail(email string) (model.User, error) {
	for _, u := range r.store.List() {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return model.User{}, &model.UserNotFoundError{} // id невідомий: шукали за email
}

func (r *InMemoryUserRepository) FindAll() ([]model.User, error) {
	return r.store.List(), nil
}

func (r *InMemoryUserRepository) Update(user model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store.Get(user.ID); !ok {
		return model.User{}, &model.UserNotFoundError{ID: user.ID}
	}
	if r.emailTaken(user.Email, user.ID) {
		return model.User{}, &model.EmailTakenError{Email: user.Email}
	}
	r.store.Put(user.ID, user)
	return user, nil
}

func (r *InMemoryUserRepository) Delete(id int) error {
	if !r.store.Delete(id) {
		return &model.UserNotFoundError{ID: id}
	}
	return nil
}

func (r *InMemoryUserRepository) emailTaken(email string, exceptID int) bool {
	for _, u := range r.store.List() {
		if u.ID != exceptID && strings.EqualFold(u.Email, email) {
			return true
		}
	}
	return false
}
