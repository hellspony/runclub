-- Replace the strict UNIQUE on telegram_id with a partial unique index
-- that only enforces uniqueness for non-zero telegram_ids.
-- This allows creating members from the admin panel without a Telegram account.

-- SQLite doesn't support ALTER TABLE DROP CONSTRAINT, so we recreate the table.
-- 1. Create new table without UNIQUE on telegram_id.
CREATE TABLE IF NOT EXISTS members_new (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    fio               TEXT    NOT NULL DEFAULT '',
    telegram_username TEXT    NOT NULL DEFAULT '',
    telegram_id       INTEGER NOT NULL DEFAULT 0,
    birth_date        TEXT,
    left_at           TEXT,
    created_at        TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at        TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- 2. Copy data.
INSERT INTO members_new (id, fio, telegram_username, telegram_id, birth_date, left_at, created_at, updated_at)
    SELECT id, fio, telegram_username, telegram_id, birth_date, left_at, created_at, updated_at FROM members;

-- 3. Drop old table.
DROP TABLE members;

-- 4. Rename new table.
ALTER TABLE members_new RENAME TO members;

-- 5. Create partial unique index: only enforce uniqueness for telegram_id > 0.
CREATE UNIQUE INDEX IF NOT EXISTS idx_members_telegram_id_nonzero
    ON members (telegram_id) WHERE telegram_id != 0;
