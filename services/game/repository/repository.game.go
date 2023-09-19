package repository_game

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/lookingcoolonavespa/go_crochess_backend/domain"
	"github.com/spf13/viper"
)

type gameRepo struct {
	db *sql.DB
}

func NewGameRepo(db *sql.DB) gameRepo {
	return gameRepo{db}
}

func (c gameRepo) Get(id int) (*domain.Game, error) {
	stmt, err := c.db.Prepare(
		fmt.Sprintf(
			`SELECT * 
            FROM %s.game g
            LEFT JOIN drawrecord dr ON g.id = dr.game_id
            WHERE
                g.id = $1`,
			viper.GetString("database.schema"),
		))
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

	var game domain.Game
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
	)
	if err != nil {
		return nil, err
	}

	return &game, nil
}

func (c gameRepo) Insert(g *domain.Game) error {
	stmt, err := c.db.Prepare(fmt.Sprintf(`
    INSERT INTO %s.gameseeks (
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
		viper.GetString("database.schema")),
	)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
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

func (c gameRepo) Update(id int, version int, changes map[string]interface{}) (bool, error) {
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
    UPDATE %s.gameseeks 
    SET 
        version = %d,
        %s
    WHERE
        id = %d
    AND
        version = %d
    `,
		viper.GetString("database.schema"),
		version+1,
		updateStr,
		id,
		version,
	)
	stmt, err := c.db.Prepare(sql)
	if err != nil {
		return false, err
	}

	result, err := stmt.Exec()
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()

	return rowsAffected > 0, nil
}
