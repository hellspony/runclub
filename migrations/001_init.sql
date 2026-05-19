PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

CREATE TABLE clubs (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    name                TEXT    NOT NULL,
    telegram_chat_id    INTEGER NOT NULL UNIQUE,
    welcome_enabled     INTEGER NOT NULL DEFAULT 1,
    birthday_enabled    INTEGER NOT NULL DEFAULT 1,
    race_notify_enabled INTEGER NOT NULL DEFAULT 1,
    created_at          TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at          TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE members (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    fio               TEXT    NOT NULL DEFAULT '',
    telegram_username TEXT    NOT NULL DEFAULT '',
    telegram_id       INTEGER NOT NULL UNIQUE,
    birth_date        TEXT,
    created_at        TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at        TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE club_members (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    club_id   INTEGER NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    member_id INTEGER NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    role      TEXT    NOT NULL DEFAULT 'member' CHECK (role IN ('member','trainer','admin')),
    joined_at TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(club_id, member_id)
);
CREATE INDEX idx_club_members_club ON club_members(club_id);
CREATE INDEX idx_club_members_member ON club_members(member_id);

CREATE TABLE custom_fields (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    club_id    INTEGER NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    name       TEXT    NOT NULL,
    required   INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(club_id, name)
);

CREATE TABLE custom_field_values (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    member_id       INTEGER NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    custom_field_id INTEGER NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    value           TEXT    NOT NULL DEFAULT '',
    UNIQUE(member_id, custom_field_id)
);

CREATE TABLE locations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    club_id     INTEGER NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    address     TEXT    NOT NULL DEFAULT '',
    map_url     TEXT    NOT NULL DEFAULT '',
    description TEXT    NOT NULL DEFAULT '',
    created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_locations_club ON locations(club_id);

CREATE TABLE races (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    club_id   INTEGER NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    date      TEXT    NOT NULL,
    type      TEXT    NOT NULL DEFAULT '',
    place     TEXT    NOT NULL DEFAULT '',
    distances TEXT    NOT NULL DEFAULT '[]',
    name      TEXT    NOT NULL DEFAULT '',
    created_at TEXT   NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT   NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_races_club_date ON races(club_id, date);

CREATE TABLE race_registrations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    race_id    INTEGER NOT NULL REFERENCES races(id) ON DELETE CASCADE,
    member_id  INTEGER NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    distance   TEXT    NOT NULL DEFAULT '',
    created_at TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(race_id, member_id, distance)
);

CREATE TABLE templates (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    club_id INTEGER NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    type    TEXT    NOT NULL CHECK (type IN ('welcome','birthday','race_notify','training_new','training_done','jointrun_new')),
    name    TEXT    NOT NULL,
    content TEXT    NOT NULL,
    UNIQUE(club_id, type)
);

CREATE TABLE trainings (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    club_id      INTEGER NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    location_id  INTEGER NOT NULL REFERENCES locations(id),
    date         TEXT    NOT NULL,
    duration     INTEGER NOT NULL DEFAULT 60,
    status       TEXT    NOT NULL DEFAULT 'planned' CHECK (status IN ('planned','in_progress','confirming','completed')),
    photo_file_id TEXT   NOT NULL DEFAULT '',
    message_id   INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_trainings_club_date ON trainings(club_id, date);
CREATE INDEX idx_trainings_status ON trainings(status);

CREATE TABLE training_trainers (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    training_id INTEGER NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    member_id   INTEGER NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    UNIQUE(training_id, member_id)
);

CREATE TABLE training_participants (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    training_id INTEGER NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    member_id   INTEGER NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    UNIQUE(training_id, member_id)
);

CREATE TABLE joint_runs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    club_id     INTEGER NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    location_id INTEGER NOT NULL REFERENCES locations(id),
    creator_id  INTEGER NOT NULL REFERENCES members(id),
    date        TEXT    NOT NULL,
    message_id  INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_joint_runs_club_date ON joint_runs(club_id, date);

CREATE TABLE joint_run_participants (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    joint_run_id INTEGER NOT NULL REFERENCES joint_runs(id) ON DELETE CASCADE,
    member_id    INTEGER NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    UNIQUE(joint_run_id, member_id)
);

CREATE TABLE bot_states (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER NOT NULL,
    chat_id     INTEGER NOT NULL,
    flow        TEXT    NOT NULL,
    step        INTEGER NOT NULL DEFAULT 0,
    payload     TEXT    NOT NULL DEFAULT '{}',
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(telegram_id, chat_id, flow)
);

CREATE TABLE race_notification_log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    club_id    INTEGER NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    race_id    INTEGER NOT NULL REFERENCES races(id) ON DELETE CASCADE,
    sent_date  TEXT    NOT NULL,
    UNIQUE(club_id, race_id, sent_date)
);

CREATE TABLE admin_users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);
