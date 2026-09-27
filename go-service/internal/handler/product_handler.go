// Package handler — HTTP-шар: розбір запиту, виклик сервісу, вибір статус-коду.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"

	"tech-store-go/internal/httpx"
	"tech-store-go/internal/model"
	"tech-store-go/internal/service"
)

const maxBodyBytes = 1 << 20 // 1 MB

var idPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// List: GET /products?category=Laptops -> 200
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.svc.List(r.URL.Query().Get("category")))
}

// Get: GET /products/{id} -> 200 | 400 | 404
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	p, err := h.svc.Get(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

// Create: POST /products -> 201 + Location | 400
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}
	p, err := h.svc.Create(in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/products/%d", p.ID))
	httpx.WriteJSON(w, http.StatusCreated, p)
}

// Update: PUT /products/{id} -> 200 | 400 | 404
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}
	p, err := h.svc.Update(id, in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

// Delete: DELETE /products/{id} -> 204 | 400 | 404
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseID читає {id} зі шляху. Якщо він некоректний — відповідає 400 і повертає false.
func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if !idPattern.MatchString(raw) || err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid product id: must be a positive integer")
		return 0, false
	}
	return id, true
}

// decodeInput читає JSON-тіло. Порожнє тіло не є помилкою розбору — його
// відхилить валідація (400 з переліком обов'язкових полів).
func decodeInput(w http.ResponseWriter, r *http.Request) (model.ProductInput, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var in model.ProductInput
	err := json.NewDecoder(r.Body).Decode(&in)
	if err == nil || errors.Is(err, io.EOF) {
		return in, true
	}

	var (
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
		tooBig    *http.MaxBytesError
	)
	switch {
	case errors.As(err, &tooBig):
		httpx.WriteError(w, http.StatusRequestEntityTooLarge, "Request body too large")
	case errors.As(err, &typeErr):
		if typeErr.Field == "" {
			httpx.WriteError(w, http.StatusBadRequest, "Validation failed", "request body must be a JSON object")
		} else {
			httpx.WriteError(w, http.StatusBadRequest, "Validation failed",
				fmt.Sprintf("field %q has invalid type (expected %s)", typeErr.Field, jsonTypeName(typeErr.Type)))
		}
	case errors.As(err, &syntaxErr), errors.Is(err, io.ErrUnexpectedEOF):
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON in request body")
	default:
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request body")
	}
	return in, false
}

// jsonTypeName перетворює Go-тип на зрозумілу назву JSON-типу для повідомлення про помилку.
func jsonTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int64:
		return "integer"
	case reflect.Float64:
		return "number"
	default:
		return t.String()
	}
}

// writeServiceError переводить помилки сервісу в HTTP-статуси.
func writeServiceError(w http.ResponseWriter, err error) {
	var validationErr *service.ValidationError
	switch {
	case errors.Is(err, service.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "Product not found")
	case errors.As(err, &validationErr):
		httpx.WriteError(w, http.StatusBadRequest, "Validation failed", validationErr.Details...)
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
	}
}
