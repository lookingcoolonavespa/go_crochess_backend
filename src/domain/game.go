package domain

import (
	"context"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/utils"
)

type Game struct {
	ID                   int    `json:"id"`
	WhiteID              string `json:"white_id"`
	BlackID              string `json:"black_id"`
	Time                 int    `json:"time"`
	Increment            int    `json:"increment"`
	TimeStampAtTurnStart int64  `json:"time_stamp_at_turn_start"`
	WhiteTime            int    `json:"white_time"`
	BlackTime            int    `json:"black_time"`
	History              string `json:"history"`
	Moves                string `json:"moves"`
	Result               string `json:"result"`
	Method               string `json:"method"`
	Version              int    `json:"version"`
	WhiteDrawStatus      bool   `json:"white_draw_status"`
	BlackDrawStatus      bool   `json:"black_draw_status"`
}

type GameRepo interface {
	Get(ctx context.Context, id int) (Game, error)
	Update(
		ctx context.Context,
		id int,
		version int,
		changes map[string]interface{},
	) (updated bool, err error)
	InsertAndDeleteGameseeks(
		ctx context.Context,
		g Game,
	) (gameID int, deletedGameseeks []int, err error)
}

type GameUseCase interface {
	Get(ctx context.Context, id int) (Game, error)
	UpdateOnMove(
		ctx context.Context,
		gameID int,
		playerID string,
		move string,
	) (utils.Changes, error)
}
