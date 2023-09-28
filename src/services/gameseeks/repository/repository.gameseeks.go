package repository_gameseeks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	services_database "github.com/lookingcoolonavespa/go_crochess_backend/src/services/database"
)

type gameseeksRepo struct {
	db *sql.DB
}

func (c gameseeksRepo) List(ctx context.Context) ([]domain.Gameseek, error) {
	stmt := fmt.Sprintf(`
    SELECT * FROM gameseeks`,
	)

	rows, err := c.db.QueryContext(ctx, stmt)
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
	stmt := fmt.Sprintf(`
    INSERT INTO gameseeks (
        color,
        time,
        increment,
        seeker,
    ) VALUES (
        $1, $2, $3, $4
    )`,
	)

	_, err := c.db.ExecContext(
		ctx,
		stmt,
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

func (c gameseeksRepo) Delete(ctx context.Context, db services_database.DBExecutor, seekers ...string) error {
	if len(seekers) != 2 {
		return errors.New(fmt.Sprintf("seeker count should be two\nseeker count: %d", len(seekers)))
	}
	sql := fmt.Sprintf(`
    DELETE FROM 
        gameseeks
    WHERE
        seeker 
    IN (
        $1, $2
    )`,
	)

	_, err := db.ExecContext(ctx, sql, seekers[0], seekers[1])
	if err != nil {
		return err
	}

	return nil
}

func NewGameseeksRepo(db *sql.DB) gameseeksRepo {
	return gameseeksRepo{db}
}
