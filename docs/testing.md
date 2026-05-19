# Testing Conventions

## Framework

- **testify** (`assert` + `require`) for assertions
- **gomock** (`go.uber.org/mock`) for generated mocks

## Use Case Tests

`internal/usecase/*_test.go`

- Use generated gomock mocks from `internal/mocks/`.
- Pattern: `ctrl := gomock.NewController(t)` + `defer ctrl.Finish()`.
- Configure mock behavior with `repo.EXPECT().Method(gomock.Any(), args).Return(results)`.
- For error returns, use `assert.AnError` as a sentinel.

## Repository Tests

`internal/repository/sqlite/*_test.go`

- Real SQLite in `t.TempDir()` via `setupTestDB` helper.
- No mocks needed — test against actual database.
- Use `mustCreateClub`, `mustCreateMember`, `mustCreateLocation` helpers for test data.

## Assertions

- `require.NoError(t, err)` — test must stop on failure (e.g. setup, critical path).
- `require.Error(t, err)` — test must stop on failure when expecting an error.
- `assert.Equal(t, expected, actual)` — non-fatal comparison, test continues.
- `assert.Len`, `assert.Empty`, `assert.True`, etc. — for non-critical checks.
- **testifylint** is enabled: use `require` for error assertions when the test can't continue.

## Mock Generation

- Mocks are generated from repository interfaces via `mockgen` (`make generate`).
- `go:generate` directives live in `internal/domain/repository/generate.go`.
- Generated files are committed to `internal/mocks/`.
- After adding a new repository interface or method: run `make generate` to update mocks.

## Running

- `make test` — runs all tests with `-race` flag and coverage.
- `make lint` — includes `testifylint` checks.
