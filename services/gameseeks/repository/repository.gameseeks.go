package repository_gameseeks

import (
	"database/sql"
	"fmt"

	"github.com/lookingcoolonavespa/go_crochess_backend/domain"
	"github.com/spf13/viper"
)

type gameseeksRepo struct {
	db *sql.DB
}

func (c gameseeksRepo) List() ([]domain.Gameseek, error) {
	stmt, err := c.db.Prepare(
		fmt.Sprintf(`
    SELECT * FROM %s.gameseeks`,
			viper.GetString("database.schema")))
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var gameseeks []domain.Gameseek
	for rows.Next() {
		var gameseek domain.Gameseek

		err := rows.Scan(
			&gameseek.ID,
			&gameseek.Color,
			&gameseek.Time,
			&gameseek.Increment,
			&gameseek.Seeker,
		)
		if err != nil {
			return nil, err
		}

		gameseeks = append(gameseeks, gameseek)
	}

	return gameseeks, nil
}

func (c gameseeksRepo) Insert(gs *domain.Gameseek) error {
	stmt, err := c.db.Prepare(fmt.Sprintf(`
    INSERT INTO %s.gameseeks (
        color,
        time,
        increment,
        seeker,
    ) VALUES (
        $1, $2, $3, $4
    )`,
		viper.GetString("database.schema")),
	)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		&gs.Color,
		&gs.Time,
		&gs.Increment,
		&gs.Seeker,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c gameseeksRepo) Delete(seekers ...string) error {
	sql := fmt.Sprintf(`
    DELETE FROM %s.gameseeks
    WHERE
        seeker IN (`,
		viper.GetString("database.schema"),
	)

	for i, s := range seekers {
		if i == len(seekers)-1 {
			sql += fmt.Sprintf("'%s'", s)
		} else {
			sql += fmt.Sprintf("'%s', ", s)
		}
	}
	sql += ")"

	stmt, err := c.db.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return nil
}

func NewGameseeksRepo(db *sql.DB) gameseeksRepo {
	return gameseeksRepo{db}
}
