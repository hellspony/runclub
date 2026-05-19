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

## Topic-Specific Conventions

- **[Linting](docs/linting.md)** — linter rules, suppression policy
- **[Testing](docs/testing.md)** — testify, gomock, mock generation
- **[Database](docs/database.md)** — SQLite, migrations, patterns
- **[HTTP API](docs/http-api.md)** — routing, auth, response format
- **[Frontend](docs/frontend.md)** — Vue 3, field naming, build
- **[Telegram Bot](docs/telegram.md)** — FSM, callbacks, escaping

## Common Pitfalls

- **`telegram_id` uniqueness**: members created from admin panel have `telegram_id = 0`. The partial unique index (`WHERE telegram_id != 0`) allows multiple zeros. When creating members from the admin panel, use `CreateMember`, not `RegisterOrGet` (which does a lookup by telegram_id).
- **`role` field**: lives in `club_members` table, not `members`. A member can have different roles in different clubs.
- **Two role systems**: `AdminRole` (`superadmin`/`admin`) for web panel auth; `MemberRole` (`member`/`trainer`/`admin`) for club membership. Don't confuse them.
- **Frontend-backend field mismatch**: always verify JSON tags match TypeScript interface fields exactly.
