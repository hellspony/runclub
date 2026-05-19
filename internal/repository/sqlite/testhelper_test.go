package sqlite_test

import (
	"testing"

	"github.com/jmoiron/sqlx"

	"runclub/internal/domain/entity"
	"runclub/internal/repository/sqlite"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return sqlx.NewDb(db, "sqlite3")
}

func mustCreateClub(t *testing.T, db *sqlx.DB, name string, telegramChatID int64) int64 {
	t.Helper()
	repo := sqlite.NewClubRepository(db)
	id, err := repo.Create(t.Context(), &entity.Club{
		Name:           name,
		TelegramChatID: telegramChatID,
	})
	if err != nil {
		t.Fatalf("create club: %v", err)
	}
	return id
}

func mustCreateMember(t *testing.T, db *sqlx.DB, fio string, telegramID int64) int64 {
	t.Helper()
	repo := sqlite.NewMemberRepository(db)
	id, err := repo.Create(t.Context(), &entity.Member{
		FIO:        fio,
		TelegramID: telegramID,
	})
	if err != nil {
		t.Fatalf("create member: %v", err)
	}
	return id
}

func mustCreateLocation(t *testing.T, db *sqlx.DB, clubID int64, name string) int64 {
	t.Helper()
	repo := sqlite.NewLocationRepository(db)
	id, err := repo.Create(t.Context(), &entity.Location{
		ClubID: clubID,
		Name:   name,
	})
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	return id
}
