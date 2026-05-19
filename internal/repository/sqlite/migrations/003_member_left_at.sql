-- Add left_at column to members (tracks when user left their last club)
ALTER TABLE members ADD COLUMN left_at TEXT;

-- Add member_cleanup_cron to config (no schema change needed, it's in config.yaml)
