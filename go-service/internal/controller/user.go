// Package controller — HTTP-шар: бере ПЕРЕВІРЕНІ дані з context (їх підготували middleware),
// викликає сервіс і формує відповідь. Помилки не форматуються тут: контролер просто повертає
// error, а httpx.Wrap перетворює його на потрібний HTTP-статус в одному центральному місці.
package controller

import (
	"fmt"
	"net/http"

	"tech-store-go/internal/httpx"
	"tech-store-go/internal/model"
)

// UserService — що контролеру потрібно від сервісу (інтерфейс на боці споживача).
type UserService interface {
	ListUsers() ([]model.User, error)
	GetUser(id int) (model.User, error)
	CreateUser(in model.UserInput) (model.User, error)
	UpdateUser(id int, in model.UserInput) (model.User, error)
	DeleteUser(id int) error
}

type UserController struct {
	svc UserService
}

func NewUserController(svc UserService) *UserController {
	return &UserController{svc: svc}
}

// List: GET /api/users -> 200
func (c *UserController) List(w http.ResponseWriter, r *http.Request) error {
	users, err := c.svc.ListUsers()
	if err != nil {
		return err
	}
	if users == nil {
		users = []model.User{} // у JSON має бути [], а не null
	}
	httpx.WriteJSON(w, http.StatusOK, users)
	return nil
}

// Get: GET /api/users/{id} -> 200 | 400 | 404
func (c *UserController) Get(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.ID(r.Context())
	if err != nil {
		return err
	}
	user, err := c.svc.GetUser(id)
	if err != nil {
		return err
	}
	httpx.WriteJSON(w, http.StatusOK, user)
	return nil
}

// Create: POST /api/users -> 201 + Location | 400 | 409
func (c *UserController) Create(w http.ResponseWriter, r *http.Request) error {
	in, err := httpx.UserInput(r.Context())
	if err != nil {
		return err
	}
	user, err := c.svc.CreateUser(in)
	if err != nil {
		return err
	}
	w.Header().Set("Location", fmt.Sprintf("/api/users/%d", user.ID))
	httpx.WriteJSON(w, http.StatusCreated, user)
	return nil
}

// Update: PUT /api/users/{id} -> 200 | 400 | 404 | 409
func (c *UserController) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.ID(r.Context())
	if err != nil {
		return err
	}
	in, err := httpx.UserInput(r.Context())
	if err != nil {
		return err
	}
	user, err := c.svc.UpdateUser(id, in)
	if err != nil {
		return err
	}
	httpx.WriteJSON(w, http.StatusOK, user)
	return nil
}

// Delete: DELETE /api/users/{id} -> 204 | 400 | 404
func (c *UserController) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.ID(r.Context())
	if err != nil {
		return err
	}
	if err := c.svc.DeleteUser(id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
