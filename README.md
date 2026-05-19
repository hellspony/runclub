# RunClub

Running club management platform with a Go backend, Telegram bot, and Vue 3 admin panel.

## Features

- **Club management** — create clubs, manage members and their roles
- **Telegram bot** — auto-register members on group join, welcome messages, training & joint run creation via conversational UI
- **Admin panel** — web UI for managing clubs, members, trainings, joint runs, races, locations, and templates
- **Scheduling** — automated birthday greetings, race notifications, training confirmations, orphan member cleanup
- **Role-based access** — superadmin (full access) and admin (per-club access) roles for the web panel

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25, Echo v4, SQLite |
| Frontend | Vue 3, TypeScript, Vite, Pinia |
| Bot | go-telegram-bot-api v5 |
| Database | SQLite (WAL mode, embedded migrations) |
| CI | GitHub Actions (lint, test, build, frontend) |

## Quick Start

### Docker (recommended)

```bash
cp .env.example .env
# Edit .env — set ADMIN_PASS, JWT_SECRET, TELEGRAM_TOKEN
docker compose up -d
```

The app is available at `http://localhost:8080`.

### Local Development

```bash
# Backend
make build
./bin/runclub

# Frontend dev server (proxies /api to localhost:8080)
cd web/admin && npm ci && npm run dev
```

## Configuration

Configuration is loaded from `config.yaml` (if present) and overridden by environment variables.

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | `8080` | HTTP server port |
| `DB_PATH` | `./data/runclub.db` | SQLite database path |
| `TELEGRAM_TOKEN` | — | Telegram bot token (bot is disabled if empty) |
| `ADMIN_USER` | `admin` | Default admin username (created on first start) |
| `ADMIN_PASS` | `changeme` | Default admin password |
| `ADMIN_ROLE` | `superadmin` | Default admin role |
| `JWT_SECRET` | `secret` | JWT signing key |
| `LOG_LEVEL` | `info` | Log level (`info`, `debug`) |
| `INIT_CLUB_NAME` | — | Auto-create a club on first start |
| `INIT_CLUB_CHAT_ID` | `0` | Telegram chat ID for the initial club |
| `INIT_ADMIN_TELEGRAM_ID` | `0` | Telegram ID of the initial club admin |
| `INIT_ADMIN_USERNAME` | — | Username of the initial club admin |

## API

All routes are under `/api/v1`. Authentication uses JWT Bearer tokens.

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/login` | No | Login, returns JWT |
| GET | `/auth/me` | Yes | Current user info |

### Clubs

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/clubs` | Yes | List clubs (filtered by role) |
| POST | `/clubs` | Superadmin | Create club |
| GET | `/clubs/:id` | Yes | Get club |
| PUT | `/clubs/:id` | Yes | Update club |
| DELETE | `/clubs/:id` | Yes | Delete club |

### Members

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/clubs/:clubId/members` | Yes | List club members |
| POST | `/clubs/:clubId/members` | Yes | Create member in club |
| PUT | `/clubs/:clubId/members/:memberId/role` | Yes | Update member role |
| GET | `/members/:id` | Yes | Get member |
| PUT | `/members/:id` | Yes | Update member |
| DELETE | `/members/:id` | Yes | Delete member |

### Trainings

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/clubs/:clubId/trainings` | Yes | List trainings |
| POST | `/clubs/:clubId/trainings` | Yes | Create training |
| GET | `/trainings/:id` | Yes | Get training |
| PUT | `/trainings/:id` | Yes | Update training |
| DELETE | `/trainings/:id` | Yes | Delete training |
| GET | `/trainings/:id/participants` | Yes | List participants |

### Joint Runs

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/clubs/:clubId/joint-runs` | Yes | List joint runs |
| POST | `/clubs/:clubId/joint-runs` | Yes | Create joint run |
| GET | `/joint-runs/:id` | Yes | Get joint run |
| DELETE | `/joint-runs/:id` | Yes | Delete joint run |
| GET | `/joint-runs/:id/participants` | Yes | List participants |

### Races

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/clubs/:clubId/races` | Yes | List races |
| POST | `/clubs/:clubId/races` | Yes | Create race |
| GET | `/races/:id` | Yes | Get race |
| PUT | `/races/:id` | Yes | Update race |
| DELETE | `/races/:id` | Yes | Delete race |
| GET | `/races/:id/registrations` | Yes | List registrations |
| POST | `/races/:id/register` | Yes | Register member |
| DELETE | `/races/:id/register` | Yes | Unregister member |

### Locations

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/clubs/:clubId/locations` | Yes | List locations |
| POST | `/clubs/:clubId/locations` | Yes | Create location |
| GET | `/locations/:id` | Yes | Get location |
| PUT | `/locations/:id` | Yes | Update location |
| DELETE | `/locations/:id` | Yes | Delete location |

### Templates

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/clubs/:clubId/templates` | Yes | List templates |
| POST | `/clubs/:clubId/templates` | Yes | Create template |
| PUT | `/templates/:id` | Yes | Update template |
| DELETE | `/templates/:id` | Yes | Delete template |

### Admin Users (superadmin only)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/admin-users` | List admin users |
| POST | `/admin-users` | Create admin user |
| DELETE | `/admin-users/:id` | Delete admin user |
| GET | `/admin-users/:id/clubs` | List assigned clubs |
| POST | `/admin-users/:id/clubs/:clubId` | Assign club |
| DELETE | `/admin-users/:id/clubs/:clubId` | Unassign club |

## Development

```bash
make lint              # Run golangci-lint
make test              # Run tests with race detector
make build             # Build binary
make frontend-build    # Build frontend
make coverage          # Generate HTML coverage report
```

## Project Structure

```
cmd/server/              Entry point and dependency wiring
internal/
  config/                Configuration loading (YAML + env)
  domain/
    entity/              Domain models (pure structs)
    repository/          Repository interfaces
  usecase/               Business logic
  delivery/
    http/                Echo HTTP handlers and middleware
    telegram/            Telegram bot handlers
  repository/sqlite/     SQLite implementations + migrations
  pkg/                   Shared utilities (tgrescape, templater)
  scheduler/             Cron jobs
web/admin/               Vue 3 admin panel
```
