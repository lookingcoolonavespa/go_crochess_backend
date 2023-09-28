package usecase_game

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	domain_timerManager "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/timerManager"
	"github.com/notnil/chess"
)

type gameUseCase struct {
	db           *sql.DB
	gameRepo     domain.GameRepo
	timerManager *domain_timerManager.TimerManager
	onTimeOut    func()
}

type Changes map[string]interface{}

func NewGameUseCase(
	db *sql.DB,
	gameRepo domain.GameRepo,
	timerManager *domain_timerManager.TimerManager,
	onTimeOut func(),
) gameUseCase {
	return gameUseCase{
		db,
		gameRepo,
		timerManager,
		onTimeOut,
	}
}

func (c gameUseCase) Get(ctx context.Context, gameID int) (*domain.Game, error) {
	game, err := c.gameRepo.Get(ctx, c.db, gameID)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func makeMove(g *domain.Game, playerID string, move string) (Changes, chess.Color, error) {
	// makeMove returns the changes that need to be made to game structured as key/value pairs,
	// the active color, and errors
	changes := make(Changes)

	gameState := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	moves := strings.Split(g.Moves, " ")

	for _, move := range moves {
		err := gameState.MoveStr(move)
		if err != nil {
			return nil, chess.NoColor, err
		}
	}

	activeColor := gameState.Position().Turn()
	if activeColor == chess.White && g.WhiteID != playerID ||
		activeColor == chess.Black && g.BlackID != playerID {
		return nil, chess.NoColor, errors.New("Invalid player.")
	}

	err := gameState.MoveStr(move)
	if err != nil {
		return nil, chess.NoColor, err
	}

	outcome := gameState.Outcome()
	if outcome != chess.NoOutcome {
		changes["Result"] = outcome.String()
		changes["Method"] = gameState.Method().String()

	} else {
		if elgibleDraw := len(gameState.EligibleDraws()) > 0; elgibleDraw {
			g.DrawRecord.Black = true
			g.DrawRecord.White = true
		}

		timeSpent := time.Now().Unix() - g.TimeStampAtTurnStart

		var activeTime int
		var fieldOfActiveTime string
		if activeColor == chess.White {
			activeTime = g.WhiteTime
			fieldOfActiveTime = "WhiteTime"
		} else {
			activeTime = g.BlackTime
			fieldOfActiveTime = "BlackTime"
		}

		base := activeTime - int(timeSpent)
		changes[fieldOfActiveTime] = base + (g.Increment * 1000)
		changes["TimeStampAtTurnStart"] = time.Now().Unix()
	}

	if len(g.Moves) > 0 {
		changes["Moves"] = g.Moves + fmt.Sprintf(" %s", move)
	} else {
		changes["Moves"] = fmt.Sprintf("%s", move)
	}

	gameState.ChangeNotation(chess.AlgebraicNotation{})
	changes["History"] = strings.TrimLeft(gameState.String(), "\n")

	return changes, activeColor.Other(), nil
}

func (c gameUseCase) handleTimer(
	ctx context.Context,
	gameID int, version int,
	duration time.Duration,
	activeColor chess.Color,
	gameOver bool,
) {
	if gameOver {
		c.timerManager.StopAndDeleteTimer(fmt.Sprint(gameID))
	} else {
		c.timerManager.StartTimer(fmt.Sprint(gameID), duration, func() {
			changes := make(map[string]interface{})
			if activeColor == chess.White {
				changes["WhiteTime"] = 0
				changes["Result"] = chess.BlackWon.String()
			} else {
				changes["BlackTime"] = 0
				changes["Result"] = chess.WhiteWon.String()
			}
			changes["Method"] = "Time out"
			c.gameRepo.Update(ctx, c.db, gameID, version, changes)
			c.onTimeOut()
		})
	}
}

func intToMillisecondsDuration(value int) time.Duration {
	return time.Duration(value) * time.Millisecond
}

func (c gameUseCase) UpdateOnMove(
	ctx context.Context,
	gameID int,
	playerID string,
	move string,
) (Changes, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	g, err := c.gameRepo.Get(ctx, tx, gameID)
	if err != nil {
		return nil, err
	}

	changes, activeColor, err := makeMove(g, playerID, move)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	updated, err := c.gameRepo.Update(ctx, tx, gameID, g.Version, changes)
	if !updated {
		tx.Rollback()
		return nil, errors.New("The move did not reach the server fast enough")
	}
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, gameOver := changes["Outcome"]
	var timerDuration time.Duration
	if activeColor == chess.White {
		timerDuration = intToMillisecondsDuration(g.WhiteTime)
	} else {
		timerDuration = intToMillisecondsDuration(g.BlackTime)
	}
	c.handleTimer(ctx, gameID, g.Version+1, timerDuration, activeColor, gameOver)

	return changes, nil
}
