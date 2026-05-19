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

func TestCreateClub(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockClubRepository(ctrl)
	uc := usecase.NewClubUseCase(repo)

	club := &entity.Club{
		Name:           "Runners",
		TelegramChatID: 12345,
	}

	repo.EXPECT().
		Create(go_mock.Any(), go_mock.Eq(club)).
		Return(int64(1), nil)

	repo.EXPECT().
		GetByID(go_mock.Any(), int64(1)).
		Return(&entity.Club{
			ID:             1,
			Name:           "Runners",
			TelegramChatID: 12345,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}, nil)

	id, err := uc.Create(context.Background(), club)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	got, err := uc.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "Runners", got.Name)
	assert.Equal(t, int64(12345), got.TelegramChatID)
}

func TestGetClubByID(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockClubRepository(ctrl)
	uc := usecase.NewClubUseCase(repo)

	t.Run("existing club", func(t *testing.T) {
		created := &entity.Club{ID: 1, Name: "Alpha", TelegramChatID: 100}
		repo.EXPECT().
			Create(go_mock.Any(), go_mock.Any()).
			Return(int64(1), nil)
		repo.EXPECT().
			GetByID(go_mock.Any(), int64(1)).
			Return(created, nil)

		uc.Create(context.Background(), &entity.Club{Name: "Alpha", TelegramChatID: 100})
		got, err := uc.GetByID(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, "Alpha", got.Name)
	})

	t.Run("non-existent club", func(t *testing.T) {
		repo.EXPECT().
			GetByID(go_mock.Any(), int64(999)).
			Return(nil, assert.AnError)

		_, err := uc.GetByID(context.Background(), 999)
		assert.Error(t, err)
	})
}

func TestListClubs(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockClubRepository(ctrl)
	uc := usecase.NewClubUseCase(repo)

	repo.EXPECT().
		List(go_mock.Any()).
		Return([]entity.Club{
			{ID: 1, Name: "A"},
			{ID: 2, Name: "B"},
		}, nil)

	clubs, err := uc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, clubs, 2)

	t.Run("empty list", func(t *testing.T) {
		emptyRepo := mocks.NewMockClubRepository(ctrl)
		emptyUC := usecase.NewClubUseCase(emptyRepo)

		emptyRepo.EXPECT().
			List(go_mock.Any()).
			Return([]entity.Club{}, nil)

		emptyClubs, listErr := emptyUC.List(context.Background())
		require.NoError(t, listErr)
		assert.Empty(t, emptyClubs)
	})
}
