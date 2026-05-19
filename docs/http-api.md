# HTTP API Conventions

## Routing

- All API routes under `/api/v1/`.
- Protected routes use `AuthMiddleware`; superadmin routes add `SuperAdminMiddleware`.
- Club-scoped sub-resources: `/clubs/:clubId/members`, `/clubs/:clubId/races`, etc.

## Request/Response

- Request/response types defined in handler files as unexported structs.
- JSON field names use `snake_case` (matching Go struct tags).
- Error responses: `{"error": "message"}`.

## Authentication

- JWT-based auth via `Authorization: Bearer <token>` header.
- Auth middleware stores `username`, `role`, `user_id` in echo context.

## Role-Based Access

- `AuthMiddleware` — any authenticated user.
- `SuperAdminMiddleware` — requires `role == "superadmin"`.
- Admin users see only clubs assigned via `admin_user_clubs` table.
