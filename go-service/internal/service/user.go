package service

import (
	"errors"

	"tech-store-go/internal/model"
)

// UserRepository — АБСТРАКЦІЯ доступу до даних, від якої залежить сервіс.
// Інтерфейс оголошено на боці споживача (service), тому сервіс нічого не знає
// про конкретну реалізацію: InMemoryUserRepository, DatabaseRepository тощо.
//
// Контракт помилок: FindByID/Update/Delete повертають *model.UserNotFoundError,
// Create/Update — *model.EmailTakenError при порушенні унікальності email.
type UserRepository interface {
	Create(user model.User) (model.User, error)
	FindByID(id int) (model.User, error)
	FindByEmail(email string) (model.User, error)
	FindAll() ([]model.User, error)
	Update(user model.User) (model.User, error)
	Delete(id int) error
}

// UserService — бізнес-логіка користувачів. Не залежить від net/http і від конкретного сховища.
type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) ListUsers() ([]model.User, error) {
	return s.repo.FindAll()
}

func (s *UserService) GetUser(id int) (model.User, error) {
	if err := checkUserID(id); err != nil {
		return model.User{}, err
	}
	return s.repo.FindByID(id)
}

func (s *UserService) CreateUser(in model.UserInput) (model.User, error) {
	data, err := prepareUserInput(in)
	if err != nil {
		return model.User{}, err
	}
	// Бізнес-правило: email унікальний (без урахування регістру).
	if _, err := s.repo.FindByEmail(data.Email); err == nil {
		return model.User{}, &model.EmailTakenError{Email: data.Email}
	} else if !errors.Is(err, model.ErrNotFound) {
		return model.User{}, err
	}
	return s.repo.Create(model.User{Name: data.Name, Email: data.Email})
}

func (s *UserService) UpdateUser(id int, in model.UserInput) (model.User, error) {
	if err := checkUserID(id); err != nil {
		return model.User{}, err
	}
	data, err := prepareUserInput(in)
	if err != nil {
		return model.User{}, err
	}
	if _, err := s.repo.FindByID(id); err != nil {
		return model.User{}, err
	}
	// Змінювати email на власний можна; на чужий — ні.
	owner, err := s.repo.FindByEmail(data.Email)
	switch {
	case err == nil && owner.ID != id:
		return model.User{}, &model.EmailTakenError{Email: data.Email}
	case err != nil && !errors.Is(err, model.ErrNotFound):
		return model.User{}, err
	}
	return s.repo.Update(model.User{ID: id, Name: data.Name, Email: data.Email})
}

func (s *UserService) DeleteUser(id int) error {
	if err := checkUserID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

// prepareUserInput: нормалізація + перевірка інваріантів. Сервіс захищає їх сам,
// навіть якщо його викликали в обхід HTTP-валідації.
func prepareUserInput(in model.UserInput) (model.UserInput, error) {
	data := model.NormalizeUserInput(in)
	if err := model.ValidateUserInput(data); err != nil {
		return model.UserInput{}, err
	}
	return data, nil
}

func checkUserID(id int) error {
	if id < 1 {
		return &model.ValidationError{Code: "invalid_id", Message: "Invalid identifier: must be a positive integer"}
	}
	return nil
}
