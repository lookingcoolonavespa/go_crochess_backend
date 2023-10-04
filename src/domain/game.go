package domain

import (
	"context"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/utils"
)

type (
	Game struct {
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

	GameRepo interface {
		Get(ctx context.Context, id int) (Game, error)
		Update(
			ctx context.Context,
			id int,
			version int,
			changes utils.Changes,
		) (updated bool, err error)
		Insert(
			ctx context.Context,
			g Game,
		) (gameID int, err error)
	}

	GameUseCase interface {
		Get(ctx context.Context, id int) (Game, error)
		UpdateOnMove(
			ctx context.Context,
			gameID int,
			playerID string,
			move string,
			onTimeOut func(utils.Changes),
		) (utils.Changes, error)
	}
)

func (g Game) IsFilledForInsert() (bool, []string) {
	missingFields := make([]string, 0)
	if g.WhiteID == "" {
		missingFields = append(missingFields, "whiteID")
	}
	if g.BlackID == "" {
		missingFields = append(missingFields, "blackID")
	}
	if g.Time == 0 {
		missingFields = append(missingFields, "whiteID")
	}

	return len(missingFields) == 0, missingFields
}
