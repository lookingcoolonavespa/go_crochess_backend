package repository_game

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bxcodec/faker"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/stretchr/testify/assert"
)

func initMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func TestGameRepo_Get(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	gameID := 0
	rows := sqlmock.NewRows([]string{
		"id",
		"white_id",
		"black_id",
		"time",
		"increment",
		"result",
		"winner",
		"version",
		"time_stamp_at_turn_start",
		"white_time",
		"black_time",
		"moves",
		"white_draw_status",
		"black_draw_status",
	}).
		AddRow(gameID, 4, 5, 5000, 0, "", "", 0, time.Now().UnixMilli(), 5000, 5000, "", false, true)

	query :=
		fmt.Sprintf(
			`SELECT *
            FROM game
            WHERE id = $1`,
		)

	mock.ExpectQuery(query).WillReturnRows(rows)

	r := NewGameRepo(db)

	game, err := r.Get(context.Background(), gameID)

	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.True(t, game.BlackDrawStatus)
}

func TestGameRepo_Update(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	mockGame := new(domain.Game)

	err := faker.FakeData(mockGame)
	assert.NoError(t, err)

	newVersion := mockGame.Version + 1
	newWhiteTime := 50000000
	move := "e2e4"

	query := fmt.Sprintf(`
    UPDATE game 
    SET 
        version = $1,
        %s = $2, %s = CONCAT(%s, $3)
    WHERE id = %d
    AND version = %d
    `,
		domain.GameWhiteTimeJsonTag,
		domain.GameMovesJsonTag,
		domain.GameMovesJsonTag,
		mockGame.ID,
		mockGame.Version,
	)

	mock.ExpectExec(query).
		WithArgs(newVersion, newWhiteTime, move).
		WillReturnResult(sqlmock.NewResult(int64(mockGame.ID), 1))

	r := NewGameRepo(db)

	changes := make(domain.GameChanges)
	changes[domain.GameWhiteTimeJsonTag] = newWhiteTime
	changes[domain.GameMovesJsonTag] = move

	assert.NoError(t, err)

	updated, err := r.Update(context.Background(), mockGame.ID, mockGame.Version, changes)

	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestGameRepo_Insert(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	insertStmt := fmt.Sprintf(`
    INSERT INTO game (
        white_id,
        black_id,
        time,
        increment,
        version,
        time_stamp_at_turn_start,
        white_time,
        black_time
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8
    ) RETURNING id`,
	)

	expectedGameID := 64

	whiteID := "4"
	blackID := "5"
	timeData := 5000
	increment := 5
	version := 1
	timeStampAtTurnStart := time.Now().UnixMilli()
	whiteTime := timeData
	blackTime := timeData

	rows := sqlmock.NewRows([]string{"id"}).AddRow(expectedGameID)

	mock.ExpectQuery(insertStmt).
		WithArgs(
			whiteID,
			blackID,
			timeData,
			increment,
			version,
			timeStampAtTurnStart,
			whiteTime,
			blackTime,
		).
		WillReturnRows(rows)

	r := NewGameRepo(db)

	gameID, err := r.Insert(
		context.Background(),
		domain.Game{
			WhiteID:              whiteID,
			BlackID:              blackID,
			Time:                 timeData,
			Increment:            increment,
			Version:              version,
			TimeStampAtTurnStart: timeStampAtTurnStart,
			WhiteTime:            whiteTime,
			BlackTime:            blackTime,
		})
	assert.NoError(t, err)

	assert.Equal(t, expectedGameID, gameID)
}
