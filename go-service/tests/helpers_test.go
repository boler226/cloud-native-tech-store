package tests

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"tech-store-go/internal/datastore"
	"tech-store-go/internal/model"
	"tech-store-go/internal/repository"
	"tech-store-go/internal/service"
)

// Перевірка ВІДПОВІДНОСТІ ІНТЕРФЕЙСУ на етапі компіляції: усі ці типи
// взаємозамінні для service.UserRepository.
var (
	_ service.UserRepository = (*repository.InMemoryUserRepository)(nil)
	_ service.UserRepository = (*sliceRepo)(nil)
	_ service.UserRepository = (*brokenRepo)(nil)
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard) // не засмічуємо вивід тестів логами запитів
	os.Exit(m.Run())
}

func newInMemoryRepo() service.UserRepository {
	return repository.NewInMemoryUserRepository(datastore.NewMemoryStore[model.User]())
}

// sliceRepo — друга, зовсім інша реалізація інтерфейсу (на зрізі, без DataStore).
// Нею доводимо, що UserService працює з будь-якою реалізацією без змін.
type sliceRepo struct {
	rows []model.User
	seq  int
}

func newSliceRepo() service.UserRepository { return &sliceRepo{} }

func (r *sliceRepo) Create(u model.User) (model.User, error) {
	r.seq++
	u.ID = r.seq
	r.rows = append(r.rows, u)
	return u, nil
}

func (r *sliceRepo) FindByID(id int) (model.User, error) {
	for _, u := range r.rows {
		if u.ID == id {
			return u, nil
		}
	}
	return model.User{}, &model.UserNotFoundError{ID: id}
}

func (r *sliceRepo) FindByEmail(email string) (model.User, error) {
	for _, u := range r.rows {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return model.User{}, &model.UserNotFoundError{}
}

func (r *sliceRepo) FindAll() ([]model.User, error) {
	return append([]model.User{}, r.rows...), nil
}

func (r *sliceRepo) Update(u model.User) (model.User, error) {
	for i := range r.rows {
		if r.rows[i].ID == u.ID {
			r.rows[i] = u
			return u, nil
		}
	}
	return model.User{}, &model.UserNotFoundError{ID: u.ID}
}

func (r *sliceRepo) Delete(id int) error {
	for i := range r.rows {
		if r.rows[i].ID == id {
			r.rows = append(r.rows[:i], r.rows[i+1:]...)
			return nil
		}
	}
	return &model.UserNotFoundError{ID: id}
}

// brokenRepo — репозиторій, який завжди падає (імітація втраченого з'єднання з БД).
type brokenRepo struct{}

var errDB = errors.New("db connection lost: secret-host:5432")

func (brokenRepo) Create(model.User) (model.User, error)  { return model.User{}, errDB }
func (brokenRepo) FindByID(int) (model.User, error)       { return model.User{}, errDB }
func (brokenRepo) FindByEmail(string) (model.User, error) { return model.User{}, errDB }
func (brokenRepo) FindAll() ([]model.User, error)         { return nil, errDB }
func (brokenRepo) Update(model.User) (model.User, error)  { return model.User{}, errDB }
func (brokenRepo) Delete(int) error                       { return errDB }
