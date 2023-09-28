package repository_game

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"time"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	services_database "github.com/lookingcoolonavespa/go_crochess_backend/src/services/database"
)

type gameRepo struct {
}

func NewGameRepo() gameRepo {
	return gameRepo{}
}

func (c gameRepo) Get(ctx context.Context, db services_database.DBExecutor, id int) (*domain.Game, error) {
	stmt :=
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
		)

	row := db.QueryRowContext(ctx, stmt, id)

	game := domain.Game{
		DrawRecord: new(domain.DrawRecord),
	}
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
		&game.DrawRecord.Black,
		&game.DrawRecord.White,
	)
	if err != nil {
		return nil, err
	}

	return &game, nil
}

func (c gameRepo) Insert(
	ctx context.Context,
	db services_database.DBExecutor,
	g *domain.Game,
) (int64, error) {
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
    )`,
	)

	res, err := db.ExecContext(
		ctx,
		gameStmt,
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
		return 0, err
	}

	gameID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	drawRecordStmt := fmt.Sprintf(`
    INSERT INTO drawrecord (
        game_id,
        white,
        black
    ) VALUES (
        $1, $2, $2
    )`,
	)

	_, err = db.ExecContext(ctx, drawRecordStmt, gameID, false)
	if err != nil {
		return 0, err
	}

	return gameID, nil
}

func (c gameRepo) Update(
	ctx context.Context,
	db services_database.DBExecutor,
	id int,
	version int,
	changes map[string]interface{},
) (bool, error) {
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

	stmt := fmt.Sprintf(`
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

	result, err := db.ExecContext(ctx, stmt)
	if err != nil {
		return false, err
	}

	fmt.Printf("%v", result)

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
