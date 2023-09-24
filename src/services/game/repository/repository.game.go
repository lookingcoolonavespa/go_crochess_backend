package repository_game

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"time"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
)

type gameRepo struct {
	db *sql.DB
}

func NewGameRepo(db *sql.DB) gameRepo {
	return gameRepo{db}
}

func (c gameRepo) Get(ctx context.Context, id int) (*domain.Game, error) {
	stmt, err := c.db.Prepare(
		fmt.Sprintf(
			`SELECT 
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
                g.id = $1`,
		))
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, id)
	if err != nil {
		return nil, err
	}

	game := domain.Game{
		DrawRecord: new(domain.DrawRecord),
	}
	rows.Next()
	err = rows.Scan(
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
		&game.DrawRecord.Black,
		&game.DrawRecord.White,
	)
	if err != nil {
		return nil, err
	}

	return &game, nil
}

func (c gameRepo) Insert(ctx context.Context, g *domain.Game) error {
	stmt, err := c.db.Prepare(fmt.Sprintf(`
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
	))
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(
		ctx,
		&g.WhiteID,
		&g.BlackID,
		&g.Time,
		&g.Increment,
		1,
		time.Now().Unix(),
		&g.Time,
		&g.Time,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c gameRepo) Update(ctx context.Context, id int, version int, changes map[string]interface{}) (bool, error) {
	var g domain.Game
	var updateStr string
	gType := reflect.TypeOf(g)
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

	sql := fmt.Sprintf(`
    UPDATE game 
    SET 
        version = %d,
        %s
    WHERE
        id = %d
    AND
        version = %d
    `,
		version+1,
		updateStr,
		id,
		version,
	)
	stmt, err := c.db.Prepare(sql)
	if err != nil {
		return false, err
	}

	result, err := stmt.ExecContext(ctx)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()

	return rowsAffected > 0, nil
}
