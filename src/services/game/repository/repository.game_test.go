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
		"black",
		"white",
	}).
		AddRow(gameID, "faa", "fab", 5000, 0, "", "", 0, time.Now().Unix(), 5000, 5000, "", "", false, true)

	query := fmt.Sprintf(`
        SELECT 
            game.*,
            drawrecord.black,
            drawrecord.white
        FROM 
            game g
        LEFT JOIN 
            drawrecord dr
        ON
            g.id = dr.game_id
        WHERE 
            g.id = $1
    `,
	)

	mock.ExpectQuery(query).WillReturnRows(rows)

	r := NewGameRepo()

	game, err := r.Get(context.Background(), db, gameID)

	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.NotNil(t, game.DrawRecord)
	assert.True(t, game.DrawRecord.White)
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
        version = %d,
        white_time = %d
    WHERE
        id = %d
    AND
        version = %d
    `,
		newVersion,
		newWhiteTime,
		gameID,
		mockGame.Version,
	)

	mock.ExpectBegin()
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(int64(gameID), 1))

	r := NewGameRepo()

	changes := make(map[string]interface{})
	changes["WhiteTime"] = newWhiteTime

	tx, err := db.Begin()
	assert.NoError(t, err)

	updated, err := r.Update(context.Background(), tx, gameID, mockGame.Version, changes)

	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestGameRepo_Insert(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	gameQuery := fmt.Sprintf(`
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

	drawrecordQuery := fmt.Sprintf(`
    INSERT INTO drawrecord (
        game_id,
        white,
        black
    ) VALUES (
        $1, $2, $2
    )`,
	)

	expectedGameID := int64(64)

	whiteID := "four"
	blackID := "five"
	timeData := 5000
	increment := 5
	version := 1
	timeStampAtTurnStart := time.Now().Unix()
	whiteTime := timeData
	blackTime := timeData

	mock.ExpectBegin()
	mock.ExpectExec(gameQuery).
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
		WillReturnResult(sqlmock.NewResult(expectedGameID, 1))
	mock.ExpectExec(drawrecordQuery).
		WithArgs(expectedGameID, false).
		WillReturnResult(sqlmock.NewResult(expectedGameID, 1))

	r := NewGameRepo()

	tx, err := db.Begin()
	assert.NoError(t, err)

	gameID, err := r.Insert(
		context.Background(),
		tx,
		&domain.Game{
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
