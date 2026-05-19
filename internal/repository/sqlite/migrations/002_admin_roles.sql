-- Add role column to admin_users (for existing databases)
-- SQLite doesn't support ADD COLUMN IF NOT EXISTS, so we use a pragma check approach.
-- This will fail silently if the column already exists in some SQLite versions,
-- so we wrap it pragmatically. Unfortunately SQLite doesn't support procedural SQL,
-- so the Go migration runner will handle the "duplicate column" error gracefully.

ALTER TABLE admin_users ADD COLUMN role TEXT NOT NULL DEFAULT 'admin';

-- Create admin_user_clubs join table
CREATE TABLE IF NOT EXISTS admin_user_clubs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_user_id INTEGER NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    club_id       INTEGER NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    UNIQUE(admin_user_id, club_id)
);

-- Upgrade first admin user to superadmin
UPDATE admin_users SET role = 'superadmin' WHERE id = (SELECT MIN(id) FROM admin_users);
