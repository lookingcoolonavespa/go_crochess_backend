package usecase_game

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/game/repository/mock"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/utils"
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

func TestGameUseCase_UpdateOnMove(t *testing.T) {
	db, mock := initMock()

	mockGameRepo := new(repository_game_mock.GameMockRepo)
	gameUseCase := NewGameUseCase(db, mockGameRepo)

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
	}

	t.Run("Success on regular move", func(t *testing.T) {
		move := "d2d4"

		changes := utils.Changes{
			"History":              "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			"TimeStampAtTurnStart": time.Now().Unix(),
			"WhiteTime":            mockGame.WhiteTime + (mockGame.Increment * 1000),
			"Moves":                fmt.Sprintf("%s %s", mockGame.Moves, move),
			"WhiteDrawStatus":      false,
			"BlackDrawStatus":      false,
		}

		mockGameRepo.On("Get", context.Background(), mockGame.ID).Return(mockGame, nil).Once()
		mockGameRepo.On("Update",
			context.Background(),
			mockGame.ID,
			mockGame.Version,
			changes,
		).
			Return(true, nil).Once()

		mock.ExpectBegin()
		_, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			move,
			func(c utils.Changes) {},
		)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Success on checkmate", func(t *testing.T) {
		mockGame2 := mockGame
		mockGame2.Moves = "f2f4 e7e5 g2g4"
		mockGame2.History = "1. f4 e5 2. g4 *"

		move := "d8h4"
		changes := utils.Changes{
			"History":              "1. f4 e5 2. g4 Qh4#  0-1",
			"Moves":                fmt.Sprintf("%s %s", mockGame2.Moves, move),
			"TimeStampAtTurnStart": time.Now().Unix(),
			"BlackTime":            mockGame.BlackTime + (mockGame.Increment * 1000),
			"Result":               chess.BlackWon.String(),
			"Method":               chess.Checkmate.String(),
			"WhiteDrawStatus":      false,
			"BlackDrawStatus":      false,
		}

		mockGameRepo.On("Get", context.Background(), mockGame2.ID).Return(mockGame2, nil).Once()
		mockGameRepo.On("Update",
			context.Background(),
			mockGame2.ID,
			mockGame2.Version,
			changes,
		).Return(true, nil).Once()

		gameUseCase := NewGameUseCase(db, mockGameRepo)

		_, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame2.ID,
			mockGame2.BlackID,
			move,
			func(utils.Changes) {},
		)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Success on fivefold repetition", func(t *testing.T) {
		mockGame2 := mockGame
		mockGame2.Moves = "e2e4 e7e5 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1"
		mockGame2.History = "1. e2e4 e7e5 2. Be2 Be7 3. Bf1 Bf8 4. Be2 Be7 5. Bf1 Bf8 6. Be2 Be7 7. Bf1 Bf8 8. Be2 Be7 9. Bf1 *"

		move := "e7f8"
		changes := utils.Changes{
			"History":              "1. e4 e5 2. Be2 Be7 3. Bf1 Bf8 4. Be2 Be7 5. Bf1 Bf8 6. Be2 Be7 7. Bf1 Bf8 8. Be2 Be7 9. Bf1 Bf8 10. Be2 Be7 11. Bf1 Bf8  1/2-1/2",
			"Moves":                fmt.Sprintf("%s %s", mockGame2.Moves, move),
			"TimeStampAtTurnStart": time.Now().Unix(),
			"BlackTime":            mockGame.BlackTime + (mockGame.Increment * 1000),
			"Result":               chess.Draw.String(),
			"Method":               chess.FivefoldRepetition.String(),
			"WhiteDrawStatus":      false,
			"BlackDrawStatus":      false,
		}

		mockGameRepo.On("Get", context.Background(), mockGame2.ID).
			Return(mockGame2, nil).
			Once()
		mockGameRepo.On("Update", context.Background(), mockGame2.ID, mockGame2.Version, changes).
			Return(true, nil).
			Once()

		_, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame2.ID,
			mockGame2.BlackID,
			move,
			func(utils.Changes) {},
		)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Success on threefold repetition", func(t *testing.T) {
		mockGame2 := mockGame
		mockGame2.Moves = "e2e4 e7e5 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1 e7f8 f1e2 f8e7 e2f1"
		mockGame2.History = "1. e2e4 e7e5 2. Be2 Be7 3. Bf1 Bf8 4. Be2 Be7 5. Bf1 Bf8 6. Be2 Be7 7. Bf1 *"

		move := "e7f8"
		changes := utils.Changes{
			"History":              "1. e4 e5 2. Be2 Be7 3. Bf1 Bf8 4. Be2 Be7 5. Bf1 Bf8 6. Be2 Be7 7. Bf1 Bf8 8. Be2 Be7 9. Bf1 Bf8  *",
			"Moves":                fmt.Sprintf("%s %s", mockGame2.Moves, move),
			"TimeStampAtTurnStart": time.Now().Unix(),
			"BlackTime":            mockGame.BlackTime + (mockGame.Increment * 1000),
			"WhiteDrawStatus":      true,
			"BlackDrawStatus":      true,
		}

		mockGameRepo.On("Get", context.Background(), mockGame2.ID).
			Return(mockGame2, nil).
			Once()
		mockGameRepo.On("Update", context.Background(), mockGame2.ID, mockGame2.Version, changes).
			Return(true, nil).
			Once()

		_, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame2.ID,
			mockGame2.BlackID,
			move,
			func(utils.Changes) {},
		)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Failed on invalid move", func(t *testing.T) {
		mockGameRepo.On("Get", context.Background(), mockGame.ID).Return(mockGame, nil).Once()

		_, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			"d4d5",
			func(utils.Changes) {},
		)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Failed on Get", func(t *testing.T) {
		mockGameRepo.On("Get", context.Background(), mockGame.ID).
			Return(domain.Game{}, errors.New("Unexpected")).Once()

		_, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			"d2d4",
			func(utils.Changes) {},
		)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Failed on Update", func(t *testing.T) {
		changes := utils.Changes{
			"History":              "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			"TimeStampAtTurnStart": time.Now().Unix(),
			"WhiteTime":            mockGame.WhiteTime + (mockGame.Increment * 1000),
			"Moves":                "e2e4 e7e5 g1f3 g8f6 d2d4",
			"WhiteDrawStatus":      false,
			"BlackDrawStatus":      false,
		}
		mockGameRepo.On("Get", context.Background(), mockGame.ID).Return(mockGame, nil).Once()
		mockGameRepo.On("Update", context.Background(), mockGame.ID, mockGame.Version, changes).
			Return(false, errors.New("Unexpected")).Once()

		changes, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			"d2d4",
			func(utils.Changes) {},
		)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Handles Timer", func(t *testing.T) {
		// run this last so waiting for the timer doesnt mess up other tests
		move := "d2d4"
		mockGame.ID = 99

		changes := utils.Changes{
			"History":              "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			"TimeStampAtTurnStart": time.Now().Unix(),
			"WhiteTime":            mockGame.WhiteTime + (mockGame.Increment * 1000),
			"Moves":                fmt.Sprintf("%s %s", mockGame.Moves, move),
			"WhiteDrawStatus":      false,
			"BlackDrawStatus":      false,
		}

		mockGameRepo.On("Get", context.Background(), mockGame.ID).Return(mockGame, nil).Once()
		mockGameRepo.On("Update", context.Background(), mockGame.ID, mockGame.Version, changes).
			Return(true, nil).Once()
		mockGameRepo.On("Update", context.Background(), mockGame.ID, mockGame.Version+1,
			utils.Changes{
				"Result":          "1-0",
				"Method":          "Time out",
				"BlackTime":       0,
				"WhiteDrawStatus": false,
				"BlackDrawStatus": false,
			},
		).Return(true, nil).Once()

		channel := make(chan string)
		gameOverMsg := "Game Over"
		gameUseCase := NewGameUseCase(
			db,
			mockGameRepo,
		)

		_, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			move,
			func(utils.Changes) { go func() { channel <- gameOverMsg }() },
		)
		assert.NoError(t, err)

		select {
		case msg := <-channel:
			assert.Equal(t, msg, gameOverMsg)
		}

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})
}
