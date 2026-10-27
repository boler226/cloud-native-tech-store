// Package app — composition root: єдине місце, де обираються конкретні реалізації
// й збирається весь ланцюжок Router -> Controller -> Service -> Repository -> DataStore.
package app

import (
	"net/http"

	"tech-store-go/internal/controller"
	"tech-store-go/internal/datastore"
	"tech-store-go/internal/handler"
	"tech-store-go/internal/model"
	"tech-store-go/internal/repository"
	"tech-store-go/internal/router"
	"tech-store-go/internal/seed"
	"tech-store-go/internal/service"
)

// Options дозволяє підмінити залежності (у тестах або при переході на БД).
type Options struct {
	// UserRepository — реалізація інтерфейсу service.UserRepository.
	// Якщо nil, використовується InMemoryUserRepository із початковими даними.
	//
	// Заміна сховища = передати сюди іншу реалізацію (напр. DatabaseRepository):
	// UserService, контролер і маршрути не змінюються.
	UserRepository service.UserRepository
}

// New збирає застосунок і повертає готовий http.Handler.
func New(opts Options) http.Handler {
	// Products (лабораторна 1)
	productHandler := handler.NewProductHandler(
		service.NewProductService(repository.NewMemoryProductRepository(seed.Products())),
	)

	// Users: DataStore -> Repository -> Service -> Controller.
	// Присвоєння змінній типу інтерфейсу — місце, де компілятор перевіряє,
	// що реалізація відповідає контракту.
	var userRepo service.UserRepository = opts.UserRepository
	if userRepo == nil {
		store := datastore.NewMemoryStore[model.User]()
		userRepo = repository.NewInMemoryUserRepository(store, seed.Users()...)
	}
	userController := controller.NewUserController(service.NewUserService(userRepo))

	return router.New(
		handler.NewHealthHandler(),
		productHandler,
		handler.NewConcurrencyHandler(service.NewConcurrencyService()),
		userController,
	)
}
