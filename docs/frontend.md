# Frontend Conventions (web/admin/)

## Stack

- Vue 3 + TypeScript + Vite + Pinia + Vue Router.

## Structure

- API client in `src/api/` — one file per resource.
- Views in `src/views/` — one file per page.
- State management in `src/stores/` — Pinia stores.

## Critical Rules

- **Field names must match backend JSON tags** (`snake_case`): `telegram_username`, not `username`.
- Vite dev server proxies `/api` to `localhost:8080`.
- Build: `vue-tsc --noEmit && vite build` (type-check then bundle).

## Member Roles

- Backend uses `"trainer"`, not `"coach"` — always match backend enum values.
