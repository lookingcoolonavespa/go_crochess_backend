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
	"github.com/bxcodec/faker"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/game/repository/mock"
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
	timeNow = func() time.Time {
		return time.Date(2023, time.October, 10, 2, 10, 10, 10, time.UTC)
	}
	db, mock := initMock()

	mockGameRepo := new(repository_game_mock.GameMockRepo)
	gameUseCase := NewGameUseCase(db, mockGameRepo)

	mockGame := domain.Game{
		ID:                   1,
		WhiteID:              "4",
		BlackID:              "5",
		Time:                 900000,
		Increment:            60,
		TimeStampAtTurnStart: timeNow().UnixMilli(),
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

		changes := domain.GameChanges{
			domain.GameHistoryJsonTag:         "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			domain.GameTimeStampJsonTag:       timeNow().UnixMilli(),
			domain.GameWhiteTimeJsonTag:       mockGame.WhiteTime + (mockGame.Increment * 1000),
			domain.GameMovesJsonTag:           fmt.Sprintf("%s %s", mockGame.Moves, move),
			domain.GameWhiteDrawStatusJsonTag: false,
			domain.GameBlackDrawStatusJsonTag: false,
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
		_, _, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			move,
			func(c domain.GameChanges) {},
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
		changes := domain.GameChanges{
			domain.GameHistoryJsonTag:         "1. f4 e5 2. g4 Qh4#  0-1",
			domain.GameMovesJsonTag:           fmt.Sprintf("%s %s", mockGame2.Moves, move),
			domain.GameTimeStampJsonTag:       timeNow().UnixMilli(),
			domain.GameBlackTimeJsonTag:       mockGame.BlackTime + (mockGame.Increment * 1000),
			domain.GameResultJsonTag:          chess.BlackWon.String(),
			domain.GameMethodJsonTag:          chess.Checkmate.String(),
			domain.GameWhiteDrawStatusJsonTag: false,
			domain.GameBlackDrawStatusJsonTag: false,
		}

		mockGameRepo.On("Get", context.Background(), mockGame2.ID).Return(mockGame2, nil).Once()
		mockGameRepo.On("Update",
			context.Background(),
			mockGame2.ID,
			mockGame2.Version,
			changes,
		).Return(true, nil).Once()

		gameUseCase := NewGameUseCase(db, mockGameRepo)

		_, _, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame2.ID,
			mockGame2.BlackID,
			move,
			func(domain.GameChanges) {},
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
		changes := domain.GameChanges{
			domain.GameHistoryJsonTag:         "1. e4 e5 2. Be2 Be7 3. Bf1 Bf8 4. Be2 Be7 5. Bf1 Bf8 6. Be2 Be7 7. Bf1 Bf8 8. Be2 Be7 9. Bf1 Bf8 10. Be2 Be7 11. Bf1 Bf8  1/2-1/2",
			domain.GameMovesJsonTag:           fmt.Sprintf("%s %s", mockGame2.Moves, move),
			domain.GameTimeStampJsonTag:       timeNow().UnixMilli(),
			domain.GameBlackTimeJsonTag:       mockGame.BlackTime + (mockGame.Increment * 1000),
			domain.GameResultJsonTag:          chess.Draw.String(),
			domain.GameMethodJsonTag:          chess.FivefoldRepetition.String(),
			domain.GameWhiteDrawStatusJsonTag: false,
			domain.GameBlackDrawStatusJsonTag: false,
		}

		mockGameRepo.On("Get", context.Background(), mockGame2.ID).
			Return(mockGame2, nil).
			Once()
		mockGameRepo.On("Update", context.Background(), mockGame2.ID, mockGame2.Version, changes).
			Return(true, nil).
			Once()

		_, _, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame2.ID,
			mockGame2.BlackID,
			move,
			func(domain.GameChanges) {},
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
		changes := domain.GameChanges{
			domain.GameHistoryJsonTag:         "1. e4 e5 2. Be2 Be7 3. Bf1 Bf8 4. Be2 Be7 5. Bf1 Bf8 6. Be2 Be7 7. Bf1 Bf8 8. Be2 Be7 9. Bf1 Bf8  *",
			domain.GameMovesJsonTag:           fmt.Sprintf("%s %s", mockGame2.Moves, move),
			domain.GameTimeStampJsonTag:       timeNow().UnixMilli(),
			domain.GameBlackTimeJsonTag:       mockGame.BlackTime + (mockGame.Increment * 1000),
			domain.GameWhiteDrawStatusJsonTag: true,
			domain.GameBlackDrawStatusJsonTag: true,
		}

		mockGameRepo.On("Get", context.Background(), mockGame2.ID).
			Return(mockGame2, nil).
			Once()
		mockGameRepo.On("Update", context.Background(), mockGame2.ID, mockGame2.Version, changes).
			Return(true, nil).
			Once()

		_, _, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame2.ID,
			mockGame2.BlackID,
			move,
			func(domain.GameChanges) {},
		)
		assert.NoError(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Failed on invalid move", func(t *testing.T) {
		mockGameRepo.On("Get", context.Background(), mockGame.ID).Return(mockGame, nil).Once()

		_, _, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			"d4d5",
			func(domain.GameChanges) {},
		)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Failed on Get", func(t *testing.T) {
		mockGameRepo.On("Get", context.Background(), mockGame.ID).
			Return(domain.Game{}, errors.New("Unexpected")).Once()

		_, _, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			"d2d4",
			func(domain.GameChanges) {},
		)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Failed on Update", func(t *testing.T) {
		changes := domain.GameChanges{
			domain.GameHistoryJsonTag:         "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			domain.GameTimeStampJsonTag:       timeNow().UnixMilli(),
			domain.GameWhiteTimeJsonTag:       mockGame.WhiteTime + (mockGame.Increment * 1000),
			domain.GameMovesJsonTag:           "e2e4 e7e5 g1f3 g8f6 d2d4",
			domain.GameWhiteDrawStatusJsonTag: false,
			domain.GameBlackDrawStatusJsonTag: false,
		}
		mockGameRepo.On("Get", context.Background(), mockGame.ID).Return(mockGame, nil).Once()
		mockGameRepo.On("Update", context.Background(), mockGame.ID, mockGame.Version, changes).
			Return(false, errors.New("Unexpected")).Once()

		changes, _, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			"d2d4",
			func(domain.GameChanges) {},
		)
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})

	t.Run("Handles Timer", func(t *testing.T) {
		move := "d2d4"
		mockGame.ID = 99

		changes := domain.GameChanges{
			domain.GameHistoryJsonTag:         "1. e4 e5 2. Nf3 Nf6 3. d4 *",
			domain.GameTimeStampJsonTag:       timeNow().UnixMilli(),
			domain.GameWhiteTimeJsonTag:       mockGame.WhiteTime + (mockGame.Increment * 1000),
			domain.GameMovesJsonTag:           fmt.Sprintf("%s %s", mockGame.Moves, move),
			domain.GameWhiteDrawStatusJsonTag: false,
			domain.GameBlackDrawStatusJsonTag: false,
		}

		mockGameRepo.On("Get", context.Background(), mockGame.ID).Return(mockGame, nil).Once()
		mockGameRepo.On("Update", context.Background(), mockGame.ID, mockGame.Version, changes).
			Return(true, nil).Once()
		mockGameRepo.On("Update", context.Background(), mockGame.ID, mockGame.Version+1,
			domain.GameChanges{
				domain.GameResultJsonTag:          "1-0",
				domain.GameMethodJsonTag:          "TimeOut",
				domain.GameBlackTimeJsonTag:       0,
				domain.GameWhiteDrawStatusJsonTag: false,
				domain.GameBlackDrawStatusJsonTag: false,
			},
		).Return(true, nil).Once()

		channel := make(chan string)
		gameOverMsg := "Game Over"
		gameUseCase := NewGameUseCase(
			db,
			mockGameRepo,
		)

		_, _, err := gameUseCase.UpdateOnMove(
			context.Background(),
			mockGame.ID,
			mockGame.WhiteID,
			move,
			func(domain.GameChanges) { go func() { channel <- gameOverMsg }() },
		)
		assert.NoError(t, err)

		timeOutChan := make(chan string)
		time.AfterFunc(time.Second*2, func() {
			timeOutChan <- "time out"
		})
		select {
		case msg := <-channel:
			assert.Equal(t, msg, gameOverMsg)
		case <-timeOutChan:
			break
		}

		mockGameRepo.AssertExpectations(t)
		gameUseCase.timerManager.StopAndDeleteTimer(mockGame.ID)
	})
}

func TestGameUseCase_OnAccept(t *testing.T) {
	timeNow = func() time.Time {
		return time.Date(2023, time.October, 10, 2, 10, 10, 10, time.UTC)
	}
	db, _ := initMock()

	mockGameRepo := new(repository_game_mock.GameMockRepo)
	gameseeksUseCase := NewGameUseCase(db, mockGameRepo)

	var mockGame domain.Game
	err := faker.FakeData(&mockGame)
	assert.NoError(t, err)

	blackID := "4"
	whiteID := "5"
	mockGame.BlackID = blackID
	mockGame.WhiteID = whiteID
	mockGame.TimeStampAtTurnStart = timeNow().UnixMilli()
	mockGame.WhiteTime = mockGame.Time
	mockGame.BlackTime = mockGame.Time

	testGameID := 65
	t.Run("Success", func(t *testing.T) {
		mockGameRepo.On("Insert", context.Background(), mockGame).
			Return(testGameID, nil).
			Once()

		gameID, err := gameseeksUseCase.OnAccept(context.Background(), mockGame, func(gc domain.GameChanges) {})
		assert.NoError(t, err)

		assert.Equal(t, testGameID, gameID)

		mockGameRepo.AssertExpectations(t)
	})

	t.Run("Failed", func(t *testing.T) {
		mockGameRepo.On("Insert", context.Background(), mockGame).
			Return(-1, errors.New("Unexpected")).
			Once()

		_, err := gameseeksUseCase.OnAccept(context.Background(), mockGame, func(gc domain.GameChanges) {})
		assert.Error(t, err)

		mockGameRepo.AssertExpectations(t)
	})
}
