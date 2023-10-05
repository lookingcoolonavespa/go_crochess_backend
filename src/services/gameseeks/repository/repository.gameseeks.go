package repository_gameseeks

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
)

type gameseeksRepo struct {
	db *sql.DB
}

func NewGameseeksRepo(db *sql.DB) gameseeksRepo {
	return gameseeksRepo{db}
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

func (c gameseeksRepo) DeleteFromSeeker(ctx context.Context, seeker string) ([]int, error) {
	sql := fmt.Sprintf(`
    DELETE FROM gameseeks
    WHERE seeker = $1
    RETURNING id
	`)
	rows, err := c.db.QueryContext(ctx, sql, seeker)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	deletedIDs := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Printf("Repository/Gameseek/DeleteFromSeeker error scanning into id: %v", err)
			return nil, err
		}
		deletedIDs = append(deletedIDs, id)
	}

	return deletedIDs, nil
}

func (c gameseeksRepo) Insert(ctx context.Context, gs domain.Gameseek) error {
	stmt := fmt.Sprintf(`
    INSERT INTO gameseeks (
        color,
        time,
        increment,
        seeker
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
