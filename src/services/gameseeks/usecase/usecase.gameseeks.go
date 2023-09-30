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
) (gameID int, deletedGameseeks []int, err error) {
	g.TimeStampAtTurnStart = time.Now().Unix()
	g.WhiteTime = g.Time
	g.BlackTime = g.Time

	gameID, deletedGameseeks, err = c.gameRepo.InsertAndDeleteGameseeks(ctx, g)
	if err != nil {
		return -1, nil, err
	}

	return gameID, deletedGameseeks, nil
}
