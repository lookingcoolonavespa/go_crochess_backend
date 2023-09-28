package domain

import (
	"context"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/database"
)

type Game struct {
	ID                   int         `json:"id"`
	WhiteID              string      `json:"white_id"`
	BlackID              string      `json:"black_id"`
	Time                 int         `json:"time"`
	Increment            int         `json:"increment"`
	TimeStampAtTurnStart int64       `json:"time_stamp_at_turn_start"`
	WhiteTime            int         `json:"white_time"`
	BlackTime            int         `json:"black_time"`
	History              string      `json:"history"`
	Moves                string      `json:"moves"`
	Result               string      `json:"result"`
	Method               string      `json:"method"`
	Version              int         `json:"version"`
	DrawRecord           *DrawRecord `json:"draw_record"`
}

type GameRepo interface {
	Get(ctx context.Context, db services_database.DBExecutor, id int) (*Game, error)
	Update(ctx context.Context, db services_database.DBExecutor, id int, version int, changes map[string]interface{}) (bool, error)
	Insert(ctx context.Context, db services_database.DBExecutor, g *Game) (int64, error)
}

type GameUseCase interface {
	Get(ctx context.Context, id int) (*Game, error)
	Start(context.Context, *Game) (int64, error)
	UpdateOnMove(ctx context.Context, gameID int, playerID string, move string) error
}
