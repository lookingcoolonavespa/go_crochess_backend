package usecase_gameseeks

import (
	"context"
	"database/sql"
	"time"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
)

type gameseeksUseCase struct {
	db       *sql.DB
	gameRepo domain.GameRepo
}

func NewGameseeksUseCase(
	db *sql.DB,
	gameRepo domain.GameRepo,
) gameseeksUseCase {
	return gameseeksUseCase{
		db,
		gameRepo,
	}
}

func (c gameseeksUseCase) OnAccept(
	ctx context.Context,
	g domain.Game,
) (gameID int, err error) {
	g.TimeStampAtTurnStart = time.Now().UnixMilli()
	g.WhiteTime = g.Time
	g.BlackTime = g.Time

	gameID, err = c.gameRepo.Insert(ctx, g)
	if err != nil {
		return -1, err
	}

	return gameID, nil
}
