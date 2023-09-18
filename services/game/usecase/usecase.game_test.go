package usecase_game

import (
	"database/sql"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bxcodec/faker"
	"github.com/lookingcoolonavespa/go_crochess_backend/domain"
	repository_game_mock "github.com/lookingcoolonavespa/go_crochess_backend/services/game/repository/mock"
	repository_gameseeks_mock "github.com/lookingcoolonavespa/go_crochess_backend/services/gameseeks/repository/mock"
	"github.com/stretchr/testify/assert"
)

func initMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func TestGameUseCase_Insert(t *testing.T) {
	mockGameRepo := new(repository_game_mock.GameMockRepo)
	mockGameseeksRepo := new(repository_gameseeks_mock.GameseeksMockRepo)

	var mockGame domain.Game
	err := faker.FakeData(&mockGame)
	assert.NoError(t, err)

	blackID := "blackid"
	whiteID := "whiteid"
	mockGame.BlackID = blackID
	mockGame.WhiteID = whiteID

	t.Run("Success", func(t *testing.T) {
		mockGameRepo.On("Insert", &mockGame).Return(nil).Once()
		mockGameseeksRepo.On("Delete", mockGame.WhiteID, mockGame.BlackID).Return(nil).Once()

		gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo)

		err := gameUseCase.Insert(&mockGame)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
		mockGameseeksRepo.AssertExpectations(t)
	})

	t.Run("Failed on Insert", func(t *testing.T) {
		mockGameRepo.On("Insert", &mockGame).Return(errors.New("Unexpected")).Once()

		gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo)

		err := gameUseCase.Insert(&mockGame)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		mockGameseeksRepo.AssertExpectations(t)
	})
	t.Run("Failed on Delete", func(t *testing.T) {
		mockGameRepo.On("Insert", &mockGame).Return(nil).Once()
		mockGameseeksRepo.On("Delete", mockGame.WhiteID, mockGame.BlackID).Return(errors.New("Unexpected")).Once()

		gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo)

		err := gameUseCase.Insert(&mockGame)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		mockGameseeksRepo.AssertExpectations(t)
	})
}

func TestGameUseCase_UpdateOnMove(t *testing.T) {
	mockGameRepo := new(repository_game_mock.GameMockRepo)
	mockGameseeksRepo := new(repository_gameseeks_mock.GameseeksMockRepo)

	mockGame := domain.Game{
		ID:                   1,
		WhiteID:              "white_player",
		BlackID:              "black_player",
		Time:                 900000,
		Increment:            60,
		TimeStampAtTurnStart: time.Now().Unix(),
		WhiteTime:            600,
		BlackTime:            600,
		History:              "1. e4 e5 2. Nf3 Nf6",
		Moves:                "e2e4 e7e5 g1f3 g8f6",
		Result:               "",
		Method:               "",
		Version:              1,
		DrawRecord:           &domain.DrawRecord{},
	}

	t.Run("Success", func(t *testing.T) {
		changes := map[string]interface{}{
			"History":              "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			"TimeStampAtTurnStart": time.Now().Unix(),
			"WhiteTime":            mockGame.WhiteTime + (mockGame.Increment * 1000),
			"Moves":                "e2e4 e7e5 g1f3 g8f6 d2d4",
		}

		mockGameRepo.On("Get", mockGame.ID).Return(&mockGame, nil).Once()
		mockGameRepo.On("Update", mockGame.ID, changes).Return(nil).Once()

		gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo)

		err := gameUseCase.UpdateOnMove(mockGame.ID, mockGame.WhiteID, "d2d4")
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
	})

	t.Run("Failed on invalid move", func(t *testing.T) {
		mockGameRepo.On("Get", mockGame.ID).Return(&mockGame, nil).Once()

		gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo)

		err := gameUseCase.UpdateOnMove(mockGame.ID, mockGame.WhiteID, "d4d5")
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
	})

	t.Run("Failed on Get", func(t *testing.T) {
		mockGameRepo.On("Get", mockGame.ID).Return(nil, errors.New("Unexpected")).Once()

		gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo)

		err := gameUseCase.UpdateOnMove(mockGame.ID, mockGame.WhiteID, "d2d4")
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
	})

	t.Run("Failed on Update", func(t *testing.T) {
		changes := map[string]interface{}{
			"History":              "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			"TimeStampAtTurnStart": time.Now().Unix(),
			"WhiteTime":            mockGame.WhiteTime + (mockGame.Increment * 1000),
			"Moves":                "e2e4 e7e5 g1f3 g8f6 d2d4",
		}
		mockGameRepo.On("Get", mockGame.ID).Return(&mockGame, nil).Once()
		mockGameRepo.On("Update", mockGame.ID, changes).Return(errors.New("Unexpected")).Once()

		gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo)

		err := gameUseCase.UpdateOnMove(mockGame.ID, mockGame.WhiteID, "d2d4")
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
	})
}
