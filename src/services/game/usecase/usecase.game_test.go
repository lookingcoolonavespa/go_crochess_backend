package usecase_game

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bxcodec/faker"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
	domain_timerManager "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/timerManager"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/game/repository/mock"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/repository/mock"
	"github.com/notnil/chess"
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
	timerManager := new(domain_timerManager.TimerManager)
	gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo, timerManager, func() {})

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

		err := gameUseCase.Insert(&mockGame)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
		mockGameseeksRepo.AssertExpectations(t)
	})

	t.Run("Failed on Insert", func(t *testing.T) {
		mockGameRepo.On("Insert", &mockGame).Return(errors.New("Unexpected")).Once()

		err := gameUseCase.Insert(&mockGame)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		mockGameseeksRepo.AssertExpectations(t)
	})

	t.Run("Failed on Delete", func(t *testing.T) {
		mockGameRepo.On("Insert", &mockGame).Return(nil).Once()
		mockGameseeksRepo.On("Delete", mockGame.WhiteID, mockGame.BlackID).Return(errors.New("Unexpected")).Once()

		err := gameUseCase.Insert(&mockGame)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		mockGameseeksRepo.AssertExpectations(t)
	})
}

func TestGameUseCase_UpdateOnMove(t *testing.T) {
	mockGameRepo := new(repository_game_mock.GameMockRepo)
	mockGameseeksRepo := new(repository_gameseeks_mock.GameseeksMockRepo)
	timerManager := new(domain_timerManager.TimerManager)
	gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo, timerManager, func() {})

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

	t.Run("Success on regular move", func(t *testing.T) {
		move := "d2d4"

		changes := map[string]interface{}{
			"History":              "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			"TimeStampAtTurnStart": time.Now().Unix(),
			"WhiteTime":            mockGame.WhiteTime + (mockGame.Increment * 1000),
			"Moves":                fmt.Sprintf("%s %s", mockGame.Moves, move),
		}

		mockGameRepo.On("Get", mockGame.ID).Return(&mockGame, nil).Once()
		mockGameRepo.On("Update", mockGame.ID, mockGame.Version, changes).Return(true, nil).Once()

		err := gameUseCase.UpdateOnMove(mockGame.ID, mockGame.WhiteID, move)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
	})

	t.Run("Success on checkmate", func(t *testing.T) {
		mockGame2 := mockGame
		mockGame2.Moves = "f2f4 e7e5 g2g4"
		mockGame2.History = "1. f4 e5 2. g4 *"

		move := "d8h4"
		changes := map[string]interface{}{
			"History": "1. f4 e5 2. g4 Qh4#  0-1",
			"Moves":   fmt.Sprintf("%s %s", mockGame2.Moves, move),
			"Result":  chess.BlackWon.String(),
			"Method":  chess.Checkmate.String(),
		}

		mockGameRepo.On("Get", mockGame2.ID).Return(&mockGame2, nil).Once()
		mockGameRepo.On("Update", mockGame2.ID, mockGame2.Version, changes).Return(true, nil).Once()

		gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo, timerManager, func() {})

		err := gameUseCase.UpdateOnMove(mockGame2.ID, mockGame2.BlackID, move)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
	})

	t.Run("Success on fivefold repetition", func(t *testing.T) {
		mockGame2 := mockGame
		mockGame2.Moves = "e2e4 e7e5 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1"
		mockGame2.History = "1. e2e4 e7e5 2. Be2 Be7 3. Bf1 Bf8 4. Be2 Be7 5. Bf1 Bf8 6. Be2 Be7 7. Bf1 Bf8 8. Be2 Be7 9. Bf1 *"

		move := "e7f8"
		changes := map[string]interface{}{
			"History": "1. e4 e5 2. Be2 Be7 3. Bf1 Bf8 4. Be2 Be7 5. Bf1 Bf8 6. Be2 Be7 7. Bf1 Bf8 8. Be2 Be7 9. Bf1 Bf8 10. Be2 Be7 11. Bf1 Bf8  1/2-1/2",
			"Moves":   fmt.Sprintf("%s %s", mockGame2.Moves, move),
			"Result":  chess.Draw.String(),
			"Method":  chess.FivefoldRepetition.String(),
		}

		mockGameRepo.On("Get", mockGame2.ID).Return(&mockGame2, nil).Once()
		mockGameRepo.On("Update", mockGame2.ID, mockGame2.Version, changes).Return(true, nil).Once()

		err := gameUseCase.UpdateOnMove(mockGame2.ID, mockGame2.BlackID, move)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
	})

	t.Run("Failed on invalid move", func(t *testing.T) {
		mockGameRepo.On("Get", mockGame.ID).Return(&mockGame, nil).Once()

		err := gameUseCase.UpdateOnMove(mockGame.ID, mockGame.WhiteID, "d4d5")
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
	})

	t.Run("Failed on Get", func(t *testing.T) {
		mockGameRepo.On("Get", mockGame.ID).Return(nil, errors.New("Unexpected")).Once()

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
		mockGameRepo.On("Update", mockGame.ID, mockGame.Version, changes).Return(false, errors.New("Unexpected")).Once()

		err := gameUseCase.UpdateOnMove(mockGame.ID, mockGame.WhiteID, "d2d4")
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
	})

	t.Run("Handles Timer", func(t *testing.T) {
		// run this last so waiting for the timer doesnt mess up other tests
		move := "d2d4"

		changes := map[string]interface{}{
			"History":              "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			"TimeStampAtTurnStart": time.Now().Unix(),
			"WhiteTime":            mockGame.WhiteTime + (mockGame.Increment * 1000),
			"Moves":                fmt.Sprintf("%s %s", mockGame.Moves, move),
		}

		mockGameRepo.On("Get", mockGame.ID).Return(&mockGame, nil).Once()
		mockGameRepo.On("Update", mockGame.ID, mockGame.Version, changes).Return(true, nil).Once()
		mockGameRepo.On("Update", mockGame.ID, mockGame.Version+1,
			map[string]interface{}{
				"Result":    "1-0",
				"Method":    "Time out",
				"BlackTime": 0,
			},
		).Return(true, nil).Once()

		channel := make(chan string)
		gameOverMsg := "Game Over"
		gameUseCase := NewGameUseCase(mockGameseeksRepo, mockGameRepo, timerManager, func() { channel <- gameOverMsg })

		err := gameUseCase.UpdateOnMove(mockGame.ID, mockGame.WhiteID, move)
		assert.NoError(t, err)

		select {
		case msg := <-channel:
			assert.Equal(t, msg, gameOverMsg)
		}

		mockGameRepo.AssertExpectations(t)
	})
}
