package repository_gameseeks

import (
	"context"
	"database/sql"
	"fmt"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
)

type gameseeksRepo struct {
	db *sql.DB
}

func (c gameseeksRepo) List(ctx context.Context) ([]domain.Gameseek, error) {
	stmt, err := c.db.Prepare(
		fmt.Sprintf(`
    SELECT * FROM gameseeks`,
		))
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx)
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

func (c gameseeksRepo) Insert(ctx context.Context, gs *domain.Gameseek) error {
	stmt, err := c.db.Prepare(fmt.Sprintf(`
    INSERT INTO gameseeks (
        color,
        time,
        increment,
        seeker,
    ) VALUES (
        $1, $2, $3, $4
    )`,
	))
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(
		ctx,
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

func (c gameseeksRepo) Delete(ctx context.Context, seekers ...string) error {
	sql := fmt.Sprintf(`
    DELETE FROM gameseeks
    WHERE
        seeker IN (`,
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

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func NewGameseeksRepo(db *sql.DB) gameseeksRepo {
	return gameseeksRepo{db}
}
