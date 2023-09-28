package repository_drawrecord

import (
	"context"
	"database/sql"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
)

type drawrecordRepo struct {
	db *sql.DB
}

func NewDrawRecordRepo(db *sql.DB) drawrecordRepo {
	return drawrecordRepo{db}
}

func (c drawrecordRepo) Update(ctx context.Context, tx *sql.Tx, gameID int, version, dr domain.DrawRecord) (bool, error) {
	stmt := `
    UPDATE drawrecord
    SET
        black = $1,
        white = $2,
    WHERE
        game_id = $3
    `

	res, err := tx.ExecContext(ctx, stmt, dr.Black, dr.White, gameID, version)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return false, err
	}

	return rowsAffected > 0, nil
}
