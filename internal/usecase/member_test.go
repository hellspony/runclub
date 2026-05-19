package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	go_mock "go.uber.org/mock/gomock"

	"runclub/internal/domain/entity"
	"runclub/internal/mocks"
	"runclub/internal/usecase"
)

func TestRegisterOrGet_NewMember(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	memberRepo := mocks.NewMockMemberRepository(ctrl)
	clubMemberRepo := mocks.NewMockClubMemberRepository(ctrl)
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	memberRepo.EXPECT().
		GetByTelegramID(go_mock.Any(), int64(111)).
		Return(nil, assert.AnError)
	memberRepo.EXPECT().
		Create(go_mock.Any(), go_mock.Any()).
		Return(int64(1), nil)

	member, err := uc.RegisterOrGet(context.Background(), 111, "alice")
	require.NoError(t, err)
	assert.Equal(t, int64(1), member.ID)
	assert.Equal(t, int64(111), member.TelegramID)
	assert.Equal(t, "alice", member.FIO)
	assert.Equal(t, "alice", member.TelegramUsername)
}

func TestRegisterOrGet_ExistingMember(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	memberRepo := mocks.NewMockMemberRepository(ctrl)
	clubMemberRepo := mocks.NewMockClubMemberRepository(ctrl)
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	existing := &entity.Member{ID: 1, TelegramID: 222, FIO: "bob", TelegramUsername: "bob"}
	memberRepo.EXPECT().
		GetByTelegramID(go_mock.Any(), int64(222)).
		Return(existing, nil)

	second, err := uc.RegisterOrGet(context.Background(), 222, "bob_updated")
	require.NoError(t, err)
	assert.Equal(t, int64(1), second.ID)
	assert.Equal(t, int64(222), second.TelegramID)
}

func TestAddToClub(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	memberRepo := mocks.NewMockMemberRepository(ctrl)
	clubMemberRepo := mocks.NewMockClubMemberRepository(ctrl)
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	member := &entity.Member{ID: 1, TelegramID: 333, FIO: "charlie"}
	memberRepo.EXPECT().
		GetByID(go_mock.Any(), int64(1)).
		Return(member, nil)

	clubMemberRepo.EXPECT().
		Create(go_mock.Any(), go_mock.Any()).
		Return(int64(1), nil)

	clubMemberRepo.EXPECT().
		GetByClubAndMember(go_mock.Any(), int64(10), int64(1)).
		Return(&entity.ClubMember{
			ID:       1,
			ClubID:   10,
			MemberID: 1,
			Role:     entity.RoleMember,
		}, nil)

	err := uc.AddToClub(context.Background(), 10, 1, entity.RoleMember)
	require.NoError(t, err)

	cm, err := uc.GetClubMember(context.Background(), 10, 1)
	require.NoError(t, err)
	assert.Equal(t, entity.RoleMember, cm.Role)
	assert.Equal(t, int64(10), cm.ClubID)
}

func TestUpdateRole(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	memberRepo := mocks.NewMockMemberRepository(ctrl)
	clubMemberRepo := mocks.NewMockClubMemberRepository(ctrl)
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	clubMemberRepo.EXPECT().
		UpdateRole(go_mock.Any(), int64(20), int64(1), entity.RoleTrainer).
		Return(nil)

	err := uc.UpdateRole(context.Background(), 20, 1, entity.RoleTrainer)
	require.NoError(t, err)

	t.Run("non-existent membership", func(t *testing.T) {
		clubMemberRepo.EXPECT().
			UpdateRole(go_mock.Any(), int64(999), int64(999), entity.RoleAdmin).
			Return(assert.AnError)

		updateErr := uc.UpdateRole(context.Background(), 999, 999, entity.RoleAdmin)
		assert.Error(t, updateErr)
	})
}

func TestRemoveFromClub(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	memberRepo := mocks.NewMockMemberRepository(ctrl)
	clubMemberRepo := mocks.NewMockClubMemberRepository(ctrl)
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	clubMemberRepo.EXPECT().
		Delete(go_mock.Any(), int64(10), int64(1)).
		Return(nil)
	clubMemberRepo.EXPECT().
		ListClubsByMember(go_mock.Any(), int64(1)).
		Return([]entity.ClubMember{}, nil)
	memberRepo.EXPECT().
		GetByID(go_mock.Any(), int64(1)).
		Return(&entity.Member{ID: 1, FIO: "charlie"}, nil)
	memberRepo.EXPECT().
		Update(go_mock.Any(), go_mock.Any()).
		Return(nil)

	err := uc.RemoveFromClub(context.Background(), 10, 1)
	require.NoError(t, err)
}

func TestCreateMember(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	memberRepo := mocks.NewMockMemberRepository(ctrl)
	clubMemberRepo := mocks.NewMockClubMemberRepository(ctrl)
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	memberRepo.EXPECT().
		Create(go_mock.Any(), go_mock.Any()).
		Return(int64(5), nil)

	id, err := uc.CreateMember(context.Background(), &entity.Member{FIO: "Test", TelegramID: 0})
	require.NoError(t, err)
	assert.Equal(t, int64(5), id)
}

func TestDeleteMember(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	memberRepo := mocks.NewMockMemberRepository(ctrl)
	clubMemberRepo := mocks.NewMockClubMemberRepository(ctrl)
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	memberRepo.EXPECT().
		Delete(go_mock.Any(), int64(1)).
		Return(nil)

	err := uc.DeleteMember(context.Background(), 1)
	require.NoError(t, err)
}

func TestUpdateProfile(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	memberRepo := mocks.NewMockMemberRepository(ctrl)
	clubMemberRepo := mocks.NewMockClubMemberRepository(ctrl)
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	birthDate := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	memberRepo.EXPECT().
		GetByID(go_mock.Any(), int64(1)).
		Return(&entity.Member{ID: 1, FIO: "old", TelegramUsername: "old"}, nil)
	memberRepo.EXPECT().
		Update(go_mock.Any(), go_mock.Any()).
		Return(nil)

	err := uc.UpdateProfile(context.Background(), 1, "new name", "new_username", &birthDate)
	require.NoError(t, err)
}
