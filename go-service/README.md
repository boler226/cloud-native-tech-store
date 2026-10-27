# go-service (Go, net/http)

Багаторівнева архітектура (лабораторна 4):

```
Router -> Middleware (validation) -> Controller -> Service -> Repository -> DataStore
```

| Рівень | Файли |
|---|---|
| Router | `internal/router/router.go` |
| Middleware | `internal/middleware/middleware.go` (logging, recover), `validation.go` |
| Controller | `internal/controller/user.go` |
| Service + інтерфейс `UserRepository` | `internal/service/user.go` |
| Repository (реалізація) | `internal/repository/user.go` |
| DataStore | `internal/datastore/memory.go` |
| Model + доменні помилки | `internal/model/user.go`, `errors.go` |
| Центральна обробка помилок | `internal/httpx/errors.go` |
| Composition root | `internal/app/app.go` |
| Тести | `tests/` |

```bash
go run ./cmd/server    # http://localhost:8080
go test ./...          # тести Service та API
go test -race ./...    # те саме з race detector (потрібні cgo і gcc)
```

Повний опис — у [LAB4.md](../LAB4.md).
