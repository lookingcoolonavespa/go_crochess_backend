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
		"history",
		"moves",
		"white_draw_status",
		"black_draw_status",
	}).
		AddRow(gameID, "faa", "fab", 5000, 0, "", "", 0, time.Now().Unix(), 5000, 5000, "", "", false, true)

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

	gameID := 0
	mockGame := new(domain.Game)

	err := faker.FakeData(mockGame)
	assert.NoError(t, err)

	newVersion := mockGame.Version + 1
	newWhiteTime := 50000000

	query := fmt.Sprintf(`
    UPDATE game 
    SET 
        version = $1,
        white_time = %d
    WHERE id = $2
    AND version = $3
    `,
		newWhiteTime,
	)

	mock.ExpectExec(query).
		WithArgs(newVersion, gameID, mockGame.Version).
		WillReturnResult(sqlmock.NewResult(int64(gameID), 1))

	r := NewGameRepo(db)

	changes := make(map[string]interface{})
	changes["WhiteTime"] = newWhiteTime

	assert.NoError(t, err)

	updated, err := r.Update(context.Background(), gameID, mockGame.Version, changes)

	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestGameRepo_InsertAndDeleteGameseeks(t *testing.T) {
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
    )`,
	)

	deleteQuery := fmt.Sprintf(`
    DELETE FROM gameseeks
    WHERE seeker 
    IN ( $1, $2 )
    RETURNING id`,
	)

	expectedGameID := 64

	whiteID := "four"
	blackID := "five"
	timeData := 5000
	increment := 5
	version := 1
	timeStampAtTurnStart := time.Now().Unix()
	whiteTime := timeData
	blackTime := timeData

	mock.ExpectBegin()
	mock.ExpectExec(insertStmt).
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
		WillReturnResult(sqlmock.NewResult(int64(expectedGameID), 1))

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(0).
		AddRow(1)
	mock.ExpectQuery(deleteQuery).
		WithArgs(whiteID, blackID).
		WillReturnRows(rows)

	r := NewGameRepo(db)

	gameID, deletedGameseeks, err := r.InsertAndDeleteGameseeks(
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
	assert.NotNil(t, deletedGameseeks)
	assert.Len(t, deletedGameseeks, 2)
}
