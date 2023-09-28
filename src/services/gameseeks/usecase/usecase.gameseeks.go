package usecase_gameseeks

import (
	"context"
	"database/sql"
	"time"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
)

type gameseeksUseCase struct {
	db            *sql.DB
	gameseeksRepo domain.GameseeksRepo
	gameRepo      domain.GameRepo
}

func NewGameseeksUseCase(
	db *sql.DB,
	gameseeksRepo domain.GameseeksRepo,
	gameRepo domain.GameRepo,
) gameseeksUseCase {
	return gameseeksUseCase{
		db,
		gameseeksRepo,
		gameRepo,
	}
}

func (c gameseeksUseCase) OnAccept(ctx context.Context, g *domain.Game) (int64, error) {
	g.TimeStampAtTurnStart = time.Now().Unix()
	g.WhiteTime = g.Time
	g.BlackTime = g.Time

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	gameID, err := c.gameRepo.Insert(ctx, tx, g)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	err = c.gameseeksRepo.Delete(ctx, c.db, g.WhiteID, g.BlackID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return gameID, nil
}
