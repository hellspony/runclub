package sqlite_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, member.FIO, got.FIO)
	assert.Equal(t, member.TelegramUsername, got.TelegramUsername)
	assert.Equal(t, member.TelegramID, got.TelegramID)
	require.NotNil(t, got.BirthDate)
	assert.True(t, got.BirthDate.Equal(bday))
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
	require.NoError(t, err)

	got, err := repo.GetByTelegramID(ctx, tgID)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
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
	require.NoError(t, err)
	require.Len(t, members, 2)
	names := map[string]bool{members[0].FIO: true, members[1].FIO: true}
	assert.True(t, names["Alice"])
	assert.True(t, names["Bob"])
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
	require.NoError(t, err)
	require.Len(t, trainers, 1)
	assert.Equal(t, "Coach", trainers[0].FIO)
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
	require.NoError(t, err)
	require.Len(t, members, 1)
	assert.Equal(t, "Birthday Person", members[0].FIO)
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
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := repo.GetByClubAndMember(ctx, clubID, memberID)
	require.NoError(t, err)
	assert.Equal(t, entity.RoleMember, got.Role)
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
	require.NoError(t, err)
	assert.Len(t, cms, 2)
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

	require.NoError(t, repo.UpdateRole(ctx, clubID, memberID, entity.RoleTrainer))

	got, err := repo.GetByClubAndMember(ctx, clubID, memberID)
	require.NoError(t, err)
	assert.Equal(t, entity.RoleTrainer, got.Role)
}
