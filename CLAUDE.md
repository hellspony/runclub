# RunClub — AI Assistant Instructions

## Project Overview

RunClub is a running club management platform with:
- **Go backend** (Echo HTTP API + Telegram bot + cron scheduler)
- **Vue 3 + TypeScript frontend** (admin panel)
- **SQLite** database with embedded migrations

## Build & Run Commands

```bash
make generate            # Regenerate mocks via mockgen
make build              # Build Go binary (requires CGO_ENABLED=1 for sqlite3)
make lint               # Run golangci-lint
make test               # Run tests with race detector + coverage
make frontend-build     # Install deps & build frontend
make docker-up          # Build & start container
make docker-down        # Stop container
```

After code changes, always run `make lint && make test`. After frontend changes, run `cd web/admin && npm run build`.

## Linter Rules — CRITICAL

The project uses a strict `.golangci.yml` with 60+ linters. Key rules:

- **NEVER disable linters** with `//nolint` unless you've exhausted alternatives and can explain why. If you can't fix a lint error, ask the user instead of suppressing it.
- `nolintlint` requires every `//nolint` to specify the linter name AND an explanation. Only `funlen`, `gocognit`, `golines` may omit explanation.
- **Complexity limits**: `cyclop` max 30, `gocognit` min 20, `funlen` max 120 lines / 60 statements. If a function exceeds these, refactor it — don't suppress.
- **Magic numbers** (`mnd`): use named constants instead of bare numbers. Telegram handler files are excluded from `mnd`.
- **Comments on exported symbols** must start with the symbol name: `// Member represents...` not `// This is a member...`
- **No `log` package** outside `main.go` — use `log/slog` or `go.uber.org/zap`.
- **No `math/rand`** — use `math/rand/v2`.
- **Line length** max 120 chars (enforced by `golines`).
- **Imports** sorted with `goimports` (local prefix: `runclub`).

## Architecture

Clean architecture with strict layering:

```
cmd/server/          → Entry point, wiring (main.go, wire.go)
internal/
  config/            → Configuration (YAML + env vars)
  domain/
    entity/          → Domain models (pure structs, no dependencies)
    repository/      → Repository interfaces (ports) + go:generate directives
  mocks/             → Generated gomock mocks (committed, regenerated via make generate)
  usecase/           → Business logic, depends only on domain interfaces
  delivery/
    http/            → Echo HTTP handlers + middleware
    telegram/        → Telegram bot handlers
  repository/sqlite/ → SQLite implementations of repository interfaces
  pkg/               → Shared utilities (tgrescape, templater)
  scheduler/         → Cron jobs (birthday, race notify, training confirm, cleanup)
```

### Dependency Direction

`delivery` → `usecase` → `domain` ← `repository/sqlite`

- Entities have **no imports** outside stdlib.
- Repository interfaces live in `domain/repository/`, implementations in `repository/sqlite/`.
- Use cases depend on repository **interfaces**, never on SQLite directly.
- Handlers depend on use case **interfaces**, never on repos directly.

### Adding a New Feature

1. Entity in `internal/domain/entity/`
2. Repository interface in `internal/domain/repository/`
3. SQLite implementation in `internal/repository/sqlite/`
4. Migration in `internal/repository/sqlite/migrations/` (next number: `005_*.sql`)
5. Use case interface + implementation in `internal/usecase/`
6. HTTP handler in `internal/delivery/http/`
7. Register handler in `server.go`, wire in `cmd/server/wire.go`

## Go Code Style (Uber Style Guide)

- **Error messages**: lowercase, no trailing punctuation: `fmt.Errorf("create member: %w", err)`
- **Variable naming**: short names for narrow scope (`r` for reader, `i` for loop), descriptive for wide scope
- **Interface satisfaction**: `var _ Foo = (*Bar)(nil)` for compile-time check
- **Return errors, don't panic**: only `main` may `log.Fatal`
- **No `init()`**: use explicit initialization
- **Table-driven tests**: prefer `t.Run` subtests with struct slices
- **Pointer vs value**: use pointer when modifying or when struct contains `time.Time` with zero-value semantics
- **Group `var`/`const` blocks**: related constants together, separate `iota` blocks
- **Wrap errors**: always `%w` with `fmt.Errorf` when adding context

## Testing

- **testify** (`assert` + `require`) for assertions, **gomock** (`go.uber.org/mock`) for generated mocks.
- **Repository tests** (`internal/repository/sqlite/*_test.go`): real SQLite in `t.TempDir()` via `setupTestDB`. No mocks needed.
- **Use case tests** (`internal/usecase/*_test.go`): generated gomock mocks from `internal/mocks/`. Use `gomock.NewController(t)` + `defer ctrl.Finish()`.
- **testifylint** is enabled — use `require.Error(t, err)` (not `assert.Error`) when the test must stop on failure.
- Run with `make test` (enables `-race` flag).

### Mock Generation

- Mocks are generated from repository interfaces via `mockgen` (`make generate`).
- `go:generate` directives live in `internal/domain/repository/generate.go`.
- Generated files are committed to `internal/mocks/`.
- After adding a new repository interface method: run `make generate` to update mocks.

## Database (SQLite)

- Single connection (`SetMaxOpenConns(1)`), WAL mode.
- Migrations are embedded in the binary (`internal/repository/sqlite/migrations/`), run on startup.
- **Migration naming**: `NNN_descriptive_name.sql` (sequential number).
- SQLite doesn't support `ALTER TABLE DROP CONSTRAINT` — recreate table + copy data for schema changes.
- Partial unique index pattern for nullable-like columns: `CREATE UNIQUE INDEX ... WHERE column != 0`

## HTTP API Conventions

- All API routes under `/api/v1/`.
- Protected routes use `AuthMiddleware`; superadmin routes add `SuperAdminMiddleware`.
- Request/response types defined in handler files as unexported structs.
- JSON field names use `snake_case` (matching Go struct tags).
- Error responses: `{"error": "message"}`.
- Club-scoped sub-resources: `/clubs/:clubId/members`, `/clubs/:clubId/races`, etc.

## Frontend (web/admin/)

- Vue 3 + TypeScript + Vite + Pinia + Vue Router.
- API client in `src/api/` — one file per resource.
- Views in `src/views/` — one file per page.
- **Field names must match backend JSON tags** (`snake_case`): `telegram_username`, not `username`.
- Vite dev server proxies `/api` to `localhost:8080`.
- Build: `vue-tsc --noEmit && vite build` (type-check then bundle).

## Telegram Bot

- Handlers in `internal/delivery/telegram/handlers/`.
- FSM (finite state machine) for multi-step flows stored in `bot_states` table.
- Callback data encoded as `action:payload` pairs.
- MarkdownV2 escaping via `internal/pkg/tgrescape`.

## Common Pitfalls

- **`telegram_id` uniqueness**: members created from admin panel have `telegram_id = 0`. The partial unique index (`WHERE telegram_id != 0`) allows multiple zeros. When creating members from the admin panel, use `CreateMember`, not `RegisterOrGet` (which does a lookup by telegram_id).
- **`role` field**: lives in `club_members` table, not `members`. A member can have different roles in different clubs.
- **Two role systems**: `AdminRole` (`superadmin`/`admin`) for web panel auth; `MemberRole` (`member`/`trainer`/`admin`) for club membership. Don't confuse them.
- **Frontend-backend field mismatch**: always verify JSON tags match TypeScript interface fields exactly.
