package sqlite_test

import (
	"testing"
	"time"

	"runclub/internal/domain/entity"
	"runclub/internal/repository/sqlite"
)

func TestMemberCreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMemberRepository(db)
	ctx := t.Context()

	bday := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	member := &entity.Member{
		FIO:              "Ivan Ivanov",
		TelegramUsername: "ivanov",
		TelegramID:       12345,
		BirthDate:        &bday,
	}

	id, err := repo.Create(ctx, member)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.FIO != member.FIO {
		t.Errorf("FIO: got %q, want %q", got.FIO, member.FIO)
	}
	if got.TelegramUsername != member.TelegramUsername {
		t.Errorf("TelegramUsername: got %q, want %q", got.TelegramUsername, member.TelegramUsername)
	}
	if got.TelegramID != member.TelegramID {
		t.Errorf("TelegramID: got %d, want %d", got.TelegramID, member.TelegramID)
	}
	if got.BirthDate == nil {
		t.Error("BirthDate: got nil, want non-nil")
	} else if !got.BirthDate.Equal(bday) {
		t.Errorf("BirthDate: got %v, want %v", got.BirthDate, bday)
	}
}

func TestMemberGetByTelegramID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMemberRepository(db)
	ctx := t.Context()

	tgID := int64(99999)
	id, err := repo.Create(ctx, &entity.Member{
		FIO:        "Telegram User",
		TelegramID: tgID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByTelegramID(ctx, tgID)
	if err != nil {
		t.Fatalf("GetByTelegramID: %v", err)
	}
	if got.ID != id {
		t.Errorf("ID: got %d, want %d", got.ID, id)
	}
}

func TestMemberListByClub(t *testing.T) {
	db := setupTestDB(t)
	memberRepo := sqlite.NewMemberRepository(db)
	cmRepo := sqlite.NewClubMemberRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "TestClub", -200100)

	m1, _ := memberRepo.Create(ctx, &entity.Member{FIO: "Alice", TelegramID: 3001})
	m2, _ := memberRepo.Create(ctx, &entity.Member{FIO: "Bob", TelegramID: 3002})
	memberRepo.Create(ctx, &entity.Member{FIO: "Unrelated", TelegramID: 3003})

	cmRepo.Create(ctx, &entity.ClubMember{ClubID: clubID, MemberID: m1, Role: entity.RoleMember})
	cmRepo.Create(ctx, &entity.ClubMember{ClubID: clubID, MemberID: m2, Role: entity.RoleMember})

	members, err := memberRepo.ListByClub(ctx, clubID)
	if err != nil {
		t.Fatalf("ListByClub: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	names := map[string]bool{members[0].FIO: true, members[1].FIO: true}
	if !names["Alice"] || !names["Bob"] {
		t.Errorf("expected Alice and Bob, got %q and %q", members[0].FIO, members[1].FIO)
	}
}

func TestMemberListTrainersByClub(t *testing.T) {
	db := setupTestDB(t)
	memberRepo := sqlite.NewMemberRepository(db)
	cmRepo := sqlite.NewClubMemberRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "TrainerClub", -200200)

	trainerID, _ := memberRepo.Create(ctx, &entity.Member{FIO: "Coach", TelegramID: 4001})
	memberID, _ := memberRepo.Create(ctx, &entity.Member{FIO: "Athlete", TelegramID: 4002})

	cmRepo.Create(ctx, &entity.ClubMember{ClubID: clubID, MemberID: trainerID, Role: entity.RoleTrainer})
	cmRepo.Create(ctx, &entity.ClubMember{ClubID: clubID, MemberID: memberID, Role: entity.RoleMember})

	trainers, err := memberRepo.ListTrainersByClub(ctx, clubID)
	if err != nil {
		t.Fatalf("ListTrainersByClub: %v", err)
	}
	if len(trainers) != 1 {
		t.Fatalf("expected 1 trainer, got %d", len(trainers))
	}
	if trainers[0].FIO != "Coach" {
		t.Errorf("FIO: got %q, want %q", trainers[0].FIO, "Coach")
	}
}

func TestMemberListBirthdayOn(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMemberRepository(db)
	ctx := t.Context()

	bday := time.Date(1990, 3, 25, 0, 0, 0, 0, time.UTC)
	repo.Create(ctx, &entity.Member{
		FIO:        "Birthday Person",
		TelegramID: 5001,
		BirthDate:  &bday,
	})
	repo.Create(ctx, &entity.Member{
		FIO:        "Other Person",
		TelegramID: 5002,
		BirthDate:  nil,
	})

	members, err := repo.ListBirthdayOn(ctx, 3, 25)
	if err != nil {
		t.Fatalf("ListBirthdayOn: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].FIO != "Birthday Person" {
		t.Errorf("FIO: got %q, want %q", members[0].FIO, "Birthday Person")
	}
}

func TestClubMemberCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewClubMemberRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "CMClub", -200300)
	memberID := mustCreateMember(t, db, "Member One", 6001)

	id, err := repo.Create(ctx, &entity.ClubMember{
		ClubID:   clubID,
		MemberID: memberID,
		Role:     entity.RoleMember,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := repo.GetByClubAndMember(ctx, clubID, memberID)
	if err != nil {
		t.Fatalf("GetByClubAndMember: %v", err)
	}
	if got.Role != entity.RoleMember {
		t.Errorf("Role: got %q, want %q", got.Role, entity.RoleMember)
	}
}

func TestClubMemberListClubsByMember(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewClubMemberRepository(db)
	ctx := t.Context()

	c1 := mustCreateClub(t, db, "Club1", -200410)
	c2 := mustCreateClub(t, db, "Club2", -200420)
	mID := mustCreateMember(t, db, "MultiClub Member", 6002)

	repo.Create(ctx, &entity.ClubMember{ClubID: c1, MemberID: mID, Role: entity.RoleMember})
	repo.Create(ctx, &entity.ClubMember{ClubID: c2, MemberID: mID, Role: entity.RoleTrainer})

	cms, err := repo.ListClubsByMember(ctx, mID)
	if err != nil {
		t.Fatalf("ListClubsByMember: %v", err)
	}
	if len(cms) != 2 {
		t.Fatalf("expected 2 club memberships, got %d", len(cms))
	}
}

func TestClubMemberUpdateRole(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewClubMemberRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "RoleClub", -200500)
	memberID := mustCreateMember(t, db, "Role Member", 6003)

	repo.Create(ctx, &entity.ClubMember{
		ClubID:   clubID,
		MemberID: memberID,
		Role:     entity.RoleMember,
	})

	if err := repo.UpdateRole(ctx, clubID, memberID, entity.RoleTrainer); err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}

	got, err := repo.GetByClubAndMember(ctx, clubID, memberID)
	if err != nil {
		t.Fatalf("GetByClubAndMember: %v", err)
	}
	if got.Role != entity.RoleTrainer {
		t.Errorf("Role: got %q, want %q", got.Role, entity.RoleTrainer)
	}
}
