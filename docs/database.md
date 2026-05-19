# Database Conventions (SQLite)

## Connection

- Single connection (`SetMaxOpenConns(1)`), WAL mode.

## Migrations

- Migrations are embedded in the binary (`internal/repository/sqlite/migrations/`), run on startup.
- **Migration naming**: `NNN_descriptive_name.sql` (sequential number).
- Next available number: check existing files in the migrations directory.

## SQLite Limitations

- SQLite doesn't support `ALTER TABLE DROP CONSTRAINT` — recreate table + copy data for schema changes.
- Pattern: `CREATE TABLE new`, `INSERT INTO new SELECT ... FROM old`, `DROP TABLE old`, `ALTER TABLE new RENAME TO old`.

## Patterns

- **Partial unique index** for nullable-like columns: `CREATE UNIQUE INDEX ... WHERE column != 0` — allows multiple zero values while enforcing uniqueness for non-zero.
- **Soft delete** via `left_at` column — set timestamp instead of deleting rows when membership tracking is needed.
