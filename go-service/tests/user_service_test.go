package tests

// Тести рівня Service. HTTP-сервер НЕ запускається — сервіс перевіряється напряму.
// Одні й ті самі тести проганяються на двох різних реалізаціях репозиторію.

import (
	"errors"
	"strings"
	"testing"

	"tech-store-go/internal/model"
	"tech-store-go/internal/service"
)

var john = model.UserInput{Name: "John Doe", Email: "john@example.com"}

var implementations = map[string]func() service.UserRepository{
	"InMemoryUserRepository(DataStore)": newInMemoryRepo,
	"sliceRepo(інша реалізація)":        newSliceRepo,
}

func forEachRepo(t *testing.T, fn func(t *testing.T, svc *service.UserService)) {
	t.Helper()
	for name, make := range implementations {
		t.Run(name, func(t *testing.T) { fn(t, service.NewUserService(make())) })
	}
}

func mustCreate(t *testing.T, svc *service.UserService, in model.UserInput) model.User {
	t.Helper()
	u, err := svc.CreateUser(in)
	if err != nil {
		t.Fatalf("CreateUser(%+v): %v", in, err)
	}
	return u
}

func TestService_Create(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		u := mustCreate(t, svc, john)
		if u.ID != 1 || u.Name != "John Doe" || u.Email != "john@example.com" {
			t.Fatalf("unexpected user: %+v", u)
		}
		if next := mustCreate(t, svc, model.UserInput{Name: "Jane", Email: "jane@example.com"}); next.ID != 2 {
			t.Fatalf("second user id = %d, want 2", next.ID)
		}
	})
}

func TestService_Create_NormalizesInput(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		u := mustCreate(t, svc, model.UserInput{Name: "  Ann  ", Email: "  ANN@Example.COM "})
		if u.Name != "Ann" || u.Email != "ann@example.com" {
			t.Fatalf("not normalized: %+v", u)
		}
	})
}

func TestService_Create_DuplicateEmail(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		mustCreate(t, svc, john)

		for _, email := range []string{"john@example.com", "JOHN@EXAMPLE.COM"} { // регістр не важить
			_, err := svc.CreateUser(model.UserInput{Name: "Other", Email: email})
			var taken *model.EmailTakenError
			if !errors.As(err, &taken) || !errors.Is(err, model.ErrEmailTaken) {
				t.Fatalf("email %q: want EmailTakenError, got %v", email, err)
			}
		}
	})
}

func TestService_Create_ValidationErrors(t *testing.T) {
	cases := map[string]model.UserInput{
		"empty name":      {Name: "   ", Email: "a@b.co"},
		"empty email":     {Name: "A", Email: ""},
		"bad email":       {Name: "A", Email: "not-an-email"},
		"too long name":   {Name: strings.Repeat("x", 101), Email: "a@b.co"},
		"all fields bad":  {Name: "", Email: "nope"},
		"email without .": {Name: "A", Email: "a@b"},
	}
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		for name, in := range cases {
			_, err := svc.CreateUser(in)
			var ve *model.ValidationError
			if !errors.As(err, &ve) || len(ve.Fields) == 0 {
				t.Errorf("%s: want ValidationError with fields, got %v", name, err)
			}
		}
		// Жодного некоректного користувача не збережено.
		if users, _ := svc.ListUsers(); len(users) != 0 {
			t.Fatalf("invalid input must not be stored, got %d users", len(users))
		}
	})
}

func TestService_Get(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		created := mustCreate(t, svc, john)

		got, err := svc.GetUser(created.ID)
		if err != nil || got != created {
			t.Fatalf("GetUser = %+v, %v; want %+v", got, err, created)
		}

		_, err = svc.GetUser(42)
		var nf *model.UserNotFoundError
		if !errors.As(err, &nf) || nf.ID != 42 || !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("want UserNotFoundError{42}, got %v", err)
		}
		if err.Error() != "User with id 42 was not found" {
			t.Fatalf("message = %q", err.Error())
		}
	})
}

func TestService_Get_InvalidID(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		for _, id := range []int{0, -1, -100} {
			_, err := svc.GetUser(id)
			var ve *model.ValidationError
			if !errors.As(err, &ve) || ve.Code != "invalid_id" {
				t.Errorf("id=%d: want ValidationError(invalid_id), got %v", id, err)
			}
		}
	})
}

func TestService_List(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		if users, err := svc.ListUsers(); err != nil || len(users) != 0 {
			t.Fatalf("empty list expected, got %v, %v", users, err)
		}
		mustCreate(t, svc, john)
		mustCreate(t, svc, model.UserInput{Name: "Jane", Email: "jane@example.com"})
		users, _ := svc.ListUsers()
		if len(users) != 2 || users[0].ID != 1 || users[1].ID != 2 {
			t.Fatalf("unexpected list: %+v", users)
		}
	})
}

func TestService_Update(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		u := mustCreate(t, svc, john)

		updated, err := svc.UpdateUser(u.ID, model.UserInput{Name: "Johnny", Email: "Johnny@Example.com"})
		if err != nil {
			t.Fatal(err)
		}
		if updated.ID != u.ID || updated.Name != "Johnny" || updated.Email != "johnny@example.com" {
			t.Fatalf("unexpected update result: %+v", updated)
		}
		if got, _ := svc.GetUser(u.ID); got != updated {
			t.Fatalf("update not persisted: %+v", got)
		}
	})
}

func TestService_Update_KeepOwnEmail(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		u := mustCreate(t, svc, john)
		if _, err := svc.UpdateUser(u.ID, model.UserInput{Name: "New Name", Email: "JOHN@example.com"}); err != nil {
			t.Fatalf("keeping own email must be allowed: %v", err)
		}
	})
}

func TestService_Update_Errors(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		mustCreate(t, svc, john)
		jane := mustCreate(t, svc, model.UserInput{Name: "Jane", Email: "jane@example.com"})

		// чужий email -> конфлікт
		if _, err := svc.UpdateUser(jane.ID, model.UserInput{Name: "Jane", Email: "john@example.com"}); !errors.Is(err, model.ErrEmailTaken) {
			t.Errorf("want ErrEmailTaken, got %v", err)
		}
		// неіснуючий користувач
		if _, err := svc.UpdateUser(99, john); !errors.Is(err, model.ErrNotFound) {
			t.Errorf("want ErrNotFound, got %v", err)
		}
		// некоректні дані; і дані не змінились
		var ve *model.ValidationError
		if _, err := svc.UpdateUser(jane.ID, model.UserInput{Name: "", Email: "x"}); !errors.As(err, &ve) {
			t.Errorf("want ValidationError, got %v", err)
		}
		if got, _ := svc.GetUser(jane.ID); got.Name != "Jane" {
			t.Errorf("invalid update must not change data: %+v", got)
		}
	})
}

func TestService_Delete(t *testing.T) {
	forEachRepo(t, func(t *testing.T, svc *service.UserService) {
		u := mustCreate(t, svc, john)

		if err := svc.DeleteUser(u.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.GetUser(u.ID); !errors.Is(err, model.ErrNotFound) {
			t.Errorf("deleted user must be gone, got %v", err)
		}
		if err := svc.DeleteUser(u.ID); !errors.Is(err, model.ErrNotFound) {
			t.Errorf("second delete: want ErrNotFound, got %v", err)
		}
		// email звільнився
		mustCreate(t, svc, john)
	})
}

func TestService_RepositoryFailureIsNotMaskedAsDomainError(t *testing.T) {
	svc := service.NewUserService(brokenRepo{})

	_, err := svc.GetUser(1)
	var nf *model.UserNotFoundError
	var ve *model.ValidationError
	if err == nil || errors.As(err, &nf) || errors.As(err, &ve) || !errors.Is(err, errDB) {
		t.Fatalf("storage failure must surface unchanged (-> 500), got %v", err)
	}
	if _, err := svc.CreateUser(john); !errors.Is(err, errDB) {
		t.Fatalf("CreateUser must propagate storage failure, got %v", err)
	}
}
