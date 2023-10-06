package repository_game

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/utils"
)

type gameRepo struct {
	db *sql.DB
}

func NewGameRepo(db *sql.DB) gameRepo {
	return gameRepo{db}
}

func (c gameRepo) Get(ctx context.Context, id int) (domain.Game, error) {
	query :=
		fmt.Sprintf(
			`SELECT *
            FROM game
            WHERE id = $1`,
		)

	row := c.db.QueryRowContext(ctx, query, id)

	game := domain.Game{}
	err := row.Scan(
		&game.ID,
		&game.WhiteID,
		&game.BlackID,
		&game.Time,
		&game.Increment,
		&game.Result,
		&game.Method,
		&game.Version,
		&game.TimeStampAtTurnStart,
		&game.WhiteTime,
		&game.BlackTime,
		&game.History,
		&game.Moves,
		&game.WhiteDrawStatus,
		&game.BlackDrawStatus,
	)
	if err != nil {
		log.Printf("Repo/Game/Get, error getting game: %v\n", err)
		return domain.Game{}, err
	}

	return game, nil
}

func (c gameRepo) Insert(
	ctx context.Context,
	g domain.Game,
) (gameID int, err error) {
	gameStmt := fmt.Sprintf(`
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

	rows, err := c.db.QueryContext(
		ctx,
		gameStmt,
		&g.WhiteID,
		&g.BlackID,
		&g.Time,
		&g.Increment,
		1,
		time.Now().UnixMilli(),
		&g.Time,
		&g.Time,
	)
	if err != nil {
		log.Printf("Repo/Game/Insert, error inserting game: %v\n", err)
		return 0, err
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&gameID)

	return gameID, nil
}

func (c gameRepo) Update(
	ctx context.Context,
	id int,
	version int,
	changes utils.Changes[domain.GameFieldJsonTag],
) (updated bool, err error) {
	variableCount := 1
	updatedValues := []interface{}{version + 1}
	var updateStr string
	for field, value := range changes {
		variableCount += 1
		updateStr += fmt.Sprintf("%s = $%d, ", field, variableCount)
		updatedValues = append(updatedValues, value)
	}
	// delete trailing comma and space
	updateStr = updateStr[0 : len(updateStr)-2]

	stmt := fmt.Sprintf(`
    UPDATE game 
    SET 
        version = $1,
        %s
    WHERE id = %d
    AND version = %d
    `,
		updateStr,
		id,
		version,
	)

	result, err := c.db.ExecContext(ctx, stmt, updatedValues...)
	if err != nil {
		log.Printf("Repo/Game/Updating, error updating game: sql: %s\nerr: %v\n", stmt, err)
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowsAffected != 1 {
		return false, nil
	}

	return true, nil
}
