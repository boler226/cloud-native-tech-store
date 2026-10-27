package tests

// Тести рівня API: повний ланцюжок Router -> Middleware -> Controller -> Service -> Repository
// через реальні HTTP-запити до httptest-сервера.

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"tech-store-go/internal/app"
	"tech-store-go/internal/service"
)

type response struct {
	Status  int
	Header  http.Header
	Raw     string
	Decoded any
}

type testServer struct {
	t   *testing.T
	srv *httptest.Server
}

func newTestServer(t *testing.T, repo service.UserRepository) *testServer {
	t.Helper()
	srv := httptest.NewServer(app.New(app.Options{UserRepository: repo}))
	t.Cleanup(srv.Close)
	return &testServer{t: t, srv: srv}
}

func (s *testServer) do(method, path, contentType, body string) response {
	s.t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, s.srv.URL+path, reader)
	if err != nil {
		s.t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.t.Fatal(err)
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	var decoded any
	_ = json.Unmarshal(raw, &decoded)
	return response{Status: res.StatusCode, Header: res.Header, Raw: string(raw), Decoded: decoded}
}

func (s *testServer) json(method, path, body string) response {
	s.t.Helper()
	return s.do(method, path, "application/json", body)
}

// field дістає значення верхнього рівня з JSON-об'єкта відповіді.
func (r response) field(name string) any {
	m, _ := r.Decoded.(map[string]any)
	return m[name]
}

func (r response) wantStatus(t *testing.T, want int) {
	t.Helper()
	if r.Status != want {
		t.Fatalf("status = %d, want %d; body: %s", r.Status, want, r.Raw)
	}
}

func (r response) wantError(t *testing.T, status int, code string) {
	t.Helper()
	r.wantStatus(t, status)
	if got := r.field("error"); got != code {
		t.Fatalf("error code = %v, want %q; body: %s", got, code, r.Raw)
	}
	if msg, _ := r.field("message").(string); msg == "" {
		t.Fatalf("error response must contain a message; body: %s", r.Raw)
	}
}

const johnJSON = `{"name":"John Doe","email":"john@example.com"}`

func TestAPI_Health(t *testing.T) {
	res := newTestServer(t, newInMemoryRepo()).do("GET", "/health", "", "")
	res.wantStatus(t, 200)
	if res.field("status") != "ok" {
		t.Fatalf("unexpected body: %s", res.Raw)
	}
}

func TestAPI_ListUsers(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())

	empty := s.do("GET", "/api/users", "", "")
	empty.wantStatus(t, 200)
	if strings.TrimSpace(empty.Raw) != "[]" {
		t.Fatalf("empty list must be [], got %s", empty.Raw)
	}

	s.json("POST", "/api/users", johnJSON).wantStatus(t, 201)
	list := s.do("GET", "/api/users", "", "")
	list.wantStatus(t, 200)
	if arr, _ := list.Decoded.([]any); len(arr) != 1 {
		t.Fatalf("want 1 user, got %s", list.Raw)
	}
}

func TestAPI_CreateUser(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())

	res := s.json("POST", "/api/users", johnJSON)
	res.wantStatus(t, 201)
	if loc := res.Header.Get("Location"); loc != "/api/users/1" {
		t.Fatalf("Location = %q", loc)
	}
	if res.field("id") != float64(1) || res.field("name") != "John Doe" || res.field("email") != "john@example.com" {
		t.Fatalf("unexpected body: %s", res.Raw)
	}

	norm := s.json("POST", "/api/users", `{"name":" Ann ","email":"ANN@Example.com"}`)
	norm.wantStatus(t, 201)
	if norm.field("name") != "Ann" || norm.field("email") != "ann@example.com" {
		t.Fatalf("input must be normalized: %s", norm.Raw)
	}
}

func TestAPI_CreateUser_ValidationErrors(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())

	cases := map[string]string{
		"missing fields": `{}`,
		"wrong types":    `{"name":123,"email":true}`,
		"blank name":     `{"name":"   ","email":"a@b.co"}`,
		"bad email":      `{"name":"A","email":"nope"}`,
		"null values":    `{"name":null,"email":null}`,
		"array body":     `[1,2]`,
	}
	for name, body := range cases {
		res := s.json("POST", "/api/users", body)
		res.wantError(t, 400, "validation_failed")
		if d, _ := res.field("details").([]any); len(d) == 0 {
			t.Errorf("%s: want details, got %s", name, res.Raw)
		}
	}

	// Усі помилки полів повертаються одразу, а не лише перша.
	res := s.json("POST", "/api/users", `{"name":123,"email":"nope"}`)
	details, _ := res.field("details").([]any)
	if len(details) != 2 {
		t.Fatalf("want 2 field errors, got %s", res.Raw)
	}

	// Некоректні запити не потрапили у сховище.
	if list := s.do("GET", "/api/users", "", ""); strings.TrimSpace(list.Raw) != "[]" {
		t.Fatalf("storage must stay empty, got %s", list.Raw)
	}
}

func TestAPI_CreateUser_InvalidJSON(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())
	for _, body := range []string{`{"name": `, `{oops}`, `not json`, `123`, `{"a":1} trailing`} {
		s.json("POST", "/api/users", body).wantError(t, 400, "invalid_json")
	}
}

func TestAPI_CreateUser_EmptyBodyAndContentType(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())

	s.do("POST", "/api/users", "application/json", "").wantError(t, 400, "validation_failed")
	s.do("POST", "/api/users", "text/plain", "name=John").wantError(t, 415, "unsupported_media_type")
}

func TestAPI_CreateUser_PayloadTooLarge(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())
	big := `{"name":"` + strings.Repeat("x", 2<<20) + `","email":"a@b.co"}`
	s.json("POST", "/api/users", big).wantError(t, 413, "payload_too_large")
}

func TestAPI_CreateUser_DuplicateEmail(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())
	s.json("POST", "/api/users", johnJSON).wantStatus(t, 201)

	res := s.json("POST", "/api/users", `{"name":"Other","email":"JOHN@example.com"}`)
	res.wantError(t, 409, "email_already_exists")
}

func TestAPI_GetUser(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())
	s.json("POST", "/api/users", johnJSON)

	res := s.do("GET", "/api/users/1", "", "")
	res.wantStatus(t, 200)
	if res.field("email") != "john@example.com" {
		t.Fatalf("unexpected body: %s", res.Raw)
	}
}

func TestAPI_GetUser_NotFound(t *testing.T) {
	res := newTestServer(t, newInMemoryRepo()).do("GET", "/api/users/42", "", "")
	res.wantError(t, 404, "user_not_found")
	if res.field("message") != "User with id 42 was not found" {
		t.Fatalf("message = %v", res.field("message"))
	}
}

func TestAPI_InvalidID(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())
	for _, id := range []string{"abc", "0", "-1", "1.5", "01", "99999999999999999999"} {
		s.do("GET", "/api/users/"+id, "", "").wantError(t, 400, "invalid_id")
		s.do("DELETE", "/api/users/"+id, "", "").wantError(t, 400, "invalid_id")
		s.json("PUT", "/api/users/"+id, johnJSON).wantError(t, 400, "invalid_id")
	}
}

func TestAPI_UpdateUser(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())
	s.json("POST", "/api/users", johnJSON)

	res := s.json("PUT", "/api/users/1", `{"name":"Johnny","email":"johnny@example.com"}`)
	res.wantStatus(t, 200)
	if res.field("name") != "Johnny" || res.field("id") != float64(1) {
		t.Fatalf("unexpected body: %s", res.Raw)
	}
	if got := s.do("GET", "/api/users/1", "", ""); got.field("name") != "Johnny" {
		t.Fatalf("update not persisted: %s", got.Raw)
	}
}

func TestAPI_UpdateUser_Errors(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())
	s.json("POST", "/api/users", johnJSON)
	s.json("POST", "/api/users", `{"name":"Jane","email":"jane@example.com"}`)

	s.json("PUT", "/api/users/9", johnJSON).wantError(t, 404, "user_not_found")
	s.json("PUT", "/api/users/1", `{"name":""}`).wantError(t, 400, "validation_failed")
	s.json("PUT", "/api/users/2", `{"name":"Jane","email":"john@example.com"}`).wantError(t, 409, "email_already_exists")
}

func TestAPI_DeleteUser(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())
	s.json("POST", "/api/users", johnJSON)

	del := s.do("DELETE", "/api/users/1", "", "")
	del.wantStatus(t, 204)
	if del.Raw != "" {
		t.Fatalf("204 must have empty body, got %q", del.Raw)
	}
	s.do("GET", "/api/users/1", "", "").wantError(t, 404, "user_not_found")
	s.do("DELETE", "/api/users/1", "", "").wantError(t, 404, "user_not_found")
}

func TestAPI_MethodNotAllowedAndUnknownRoute(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())

	res := s.do("PATCH", "/api/users", "", "")
	res.wantError(t, 405, "method_not_allowed")
	if res.Header.Get("Allow") != "GET, POST" {
		t.Fatalf("Allow = %q", res.Header.Get("Allow"))
	}
	one := s.do("POST", "/api/users/1", "", "")
	one.wantError(t, 405, "method_not_allowed")
	if one.Header.Get("Allow") != "GET, PUT, DELETE" {
		t.Fatalf("Allow = %q", one.Header.Get("Allow"))
	}

	s.do("GET", "/api/unknown", "", "").wantError(t, 404, "not_found")
}

func TestAPI_EveryErrorHasUnifiedFormat(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())
	codeRe := regexp.MustCompile(`^[a-z_]+$`)

	responses := []response{
		s.do("GET", "/api/users/42", "", ""),
		s.do("GET", "/api/users/abc", "", ""),
		s.json("POST", "/api/users", `{}`),
		s.do("PATCH", "/api/users", "", ""),
		s.do("GET", "/nope", "", ""),
		s.do("GET", "/products/999", "", ""), // маршрути лабораторної 1 теж у єдиному форматі
	}
	for _, r := range responses {
		code, _ := r.field("error").(string)
		msg, _ := r.field("message").(string)
		if !codeRe.MatchString(code) || msg == "" {
			t.Errorf("not in { error, message } format: %s", r.Raw)
		}
	}
}

func TestAPI_InternalServerError(t *testing.T) {
	s := newTestServer(t, brokenRepo{})

	var logs bytes.Buffer
	log.SetOutput(&logs) // помилку можна побачити в логах, але не в відповіді
	defer log.SetOutput(io.Discard)

	res := s.do("GET", "/api/users", "", "")
	res.wantError(t, 500, "internal_error")
	if strings.Contains(res.Raw, "secret-host") {
		t.Fatalf("internal details leaked to the client: %s", res.Raw)
	}
	if !strings.Contains(logs.String(), "secret-host") {
		t.Fatalf("internal error must be logged, logs: %q", logs.String())
	}
}

func TestAPI_WorksWithAnotherRepositoryImplementation(t *testing.T) {
	s := newTestServer(t, newSliceRepo()) // інша реалізація інтерфейсу — API працює так само

	created := s.json("POST", "/api/users", johnJSON)
	created.wantStatus(t, 201)
	got := s.do("GET", "/api/users/1", "", "")
	got.wantStatus(t, 200)
	if got.Raw != created.Raw {
		t.Fatalf("responses differ: %s vs %s", got.Raw, created.Raw)
	}
}

func TestAPI_DefaultAppHasSeedUsers(t *testing.T) {
	s := newTestServer(t, nil) // nil -> InMemoryUserRepository із початковими даними
	res := s.do("GET", "/api/users", "", "")
	res.wantStatus(t, 200)
	if arr, _ := res.Decoded.([]any); len(arr) != 2 {
		t.Fatalf("want 2 seed users, got %s", res.Raw)
	}
}

func TestLoggingMiddlewareFormat(t *testing.T) {
	s := newTestServer(t, newInMemoryRepo())

	var logs bytes.Buffer
	log.SetOutput(&logs)
	defer log.SetOutput(io.Discard)

	s.do("GET", "/api/users/10", "", "")

	// "2026/01/01 12:00:00 GET /api/users/10 404 0.3ms"
	re := regexp.MustCompile(`GET /api/users/10 404 \d+\.\d{1}ms\n$`)
	if !re.MatchString(logs.String()) {
		t.Fatalf("unexpected log line: %q", logs.String())
	}
}
