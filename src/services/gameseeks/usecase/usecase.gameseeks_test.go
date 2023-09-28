package usecase_gameseeks

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bxcodec/faker"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/game/repository/mock"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/repository/mock"
	"github.com/stretchr/testify/assert"
)

func initMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func TestGameUseCase_OnAccept(t *testing.T) {
	db, _ := initMock()

	mockGameRepo := new(repository_game_mock.GameMockRepo)
	mockGameseeksRepo := new(repository_gameseeks_mock.GameseeksMockRepo)
	gameseeksUseCase := NewGameseeksUseCase(db, mockGameseeksRepo, mockGameRepo)

	var mockGame domain.Game
	err := faker.FakeData(&mockGame)
	assert.NoError(t, err)

	blackID := "blackid"
	whiteID := "whiteid"
	mockGame.BlackID = blackID
	mockGame.WhiteID = whiteID

	testGameID := 65
	t.Run("Success", func(t *testing.T) {
		mockGameRepo.On("Insert", &mockGame).Return(testGameID).Once()
		mockGameseeksRepo.On("Delete", mockGame.WhiteID, mockGame.BlackID).Return(nil).Once()

		gameID, err := gameseeksUseCase.OnAccept(context.Background(), &mockGame)
		assert.NoError(t, err)

		assert.Equal(t, testGameID, gameID)

		mockGameRepo.AssertExpectations(t)
		mockGameseeksRepo.AssertExpectations(t)
	})

	t.Run("Failed on Insert", func(t *testing.T) {
		mockGameRepo.On("Insert", &mockGame).Return(errors.New("Unexpected")).Once()

		_, err := gameseeksUseCase.OnAccept(context.Background(), &mockGame)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		mockGameseeksRepo.AssertExpectations(t)
	})

	t.Run("Failed on Delete", func(t *testing.T) {
		mockGameRepo.On("Insert", &mockGame).Return(testGameID).Once()
		mockGameseeksRepo.On("Delete", mockGame.WhiteID, mockGame.BlackID).Return(errors.New("Unexpected")).Once()

		_, err := gameseeksUseCase.OnAccept(context.Background(), &mockGame)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		mockGameseeksRepo.AssertExpectations(t)
	})
}
