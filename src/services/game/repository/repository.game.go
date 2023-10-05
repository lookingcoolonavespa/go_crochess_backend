package repository_game

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
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
	changes utils.Changes,
) (updated bool, err error) {
	var game domain.Game
	var updateStr string
	gType := reflect.TypeOf(game)
	for i := 0; i < gType.NumField(); i++ {
		field := gType.Field(i)
		fieldName := field.Name

		if _, exists := changes[fieldName]; exists {
			columnName := field.Tag.Get("json")
			if columnName == "" {
				return false, errors.New(fmt.Sprintf("Encountered an error: %s is not a valid field in Game", fieldName))
			}
			updateStr += fmt.Sprintf("%s = %v, ", columnName, changes[fieldName])
		}
	}
	regex := regexp.MustCompile(`\s*,\s*$`)
	updateStr = regex.ReplaceAllString(updateStr, "")

	stmt := fmt.Sprintf(`
    UPDATE game 
    SET 
        version = $1,
        %s
    WHERE id = $2
    AND version = $3
    `,
		updateStr,
	)

	result, err := c.db.ExecContext(ctx, stmt, version+1, id, version)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowsAffected != 1 {
		return false, errors.New("Could not update game, version is invalid")
	}

	return true, nil
}
