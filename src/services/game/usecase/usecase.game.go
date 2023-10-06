package usecase_game

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	domain_timerManager "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/timerManager"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/utils"
	"github.com/notnil/chess"
)

var timeNow = time.Now

type gameUseCase struct {
	db           *sql.DB
	gameRepo     domain.GameRepo
	timerManager *domain_timerManager.TimerManager
}

func NewGameUseCase(
	db *sql.DB,
	gameRepo domain.GameRepo,
) gameUseCase {
	return gameUseCase{
		db,
		gameRepo,
		domain_timerManager.NewTimerManager(),
	}
}

func (c gameUseCase) Get(ctx context.Context, gameID int) (domain.Game, error) {
	game, err := c.gameRepo.Get(ctx, gameID)
	if err != nil {
		return domain.Game{}, err
	}

	return game, nil
}

func makeMove(
	g domain.Game,
	playerID string,
	move string,
) (utils.Changes[domain.GameFieldJsonTag], chess.Color, error) {
	// makeMove returns the changes that need to be made to game structured as key/value pairs,
	// the active color, and errors
	changes := make(utils.Changes[domain.GameFieldJsonTag])

	gameState := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	moves := strings.Split(g.Moves, " ")

	for _, m := range moves {
		if m == "" {
			break
		}
		err := gameState.MoveStr(m)
		if err != nil {
			log.Printf("Usecase/Game/makeMove, error making move to game state\nmove: %s\nerr: %v", m, err)
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
		log.Printf("Usecase/Game/makeMove, error making move to game state\nmove: %s\nerr: %v", move, err)
		return nil, chess.NoColor, err
	}

	changes[domain.GameWhiteDrawStatusJsonTag] = false
	changes[domain.GameBlackDrawStatusJsonTag] = false

	outcome := gameState.Outcome()
	if outcome != chess.NoOutcome {
		changes[domain.GameResultJsonTag] = outcome.String()
		changes[domain.GameMethodJsonTag] = gameState.Method().String()

	} else {
		if elgibleDraw := len(gameState.EligibleDraws()) > 1; elgibleDraw {
			changes[domain.GameWhiteDrawStatusJsonTag] = true
			changes[domain.GameBlackDrawStatusJsonTag] = true
		}

	}

	timeSpent := timeNow().UnixMilli() - g.TimeStampAtTurnStart

	var activeTime int
	var fieldOfActiveTime domain.GameFieldJsonTag
	if activeColor == chess.White {
		activeTime = g.WhiteTime
		fieldOfActiveTime = domain.GameWhiteTimeJsonTag
	} else {
		activeTime = g.BlackTime
		fieldOfActiveTime = domain.GameBlackTimeJsonTag
	}

	base := activeTime - int(timeSpent)
	changes[fieldOfActiveTime] = base + (g.Increment * 1000)
	changes[domain.GameTimeStampJsonTag] = timeNow().UnixMilli()

	if len(g.Moves) > 0 {
		changes[domain.GameMovesJsonTag] = g.Moves + fmt.Sprintf(" %s", move)
	} else {
		changes[domain.GameMovesJsonTag] = fmt.Sprintf("%s", move)
	}

	gameState.ChangeNotation(chess.AlgebraicNotation{})
	changes[domain.GameHistoryJsonTag] = strings.TrimLeft(gameState.String(), "\n")

	return changes, activeColor.Other(), nil
}

func (c gameUseCase) handleTimer(
	ctx context.Context,
	onTimeOut func(utils.Changes[domain.GameFieldJsonTag]),
	gameID int, version int,
	duration time.Duration,
	activeColor chess.Color,
	gameOver bool,
) {
	if gameOver {
		c.timerManager.StopAndDeleteTimer(gameID)
	} else {
		c.timerManager.StartTimer(gameID, duration, func() {
			changes := make(utils.Changes[domain.GameFieldJsonTag])
			changes[domain.GameMovesJsonTag] = "Time out"
			changes[domain.GameWhiteDrawStatusJsonTag] = false
			changes[domain.GameBlackDrawStatusJsonTag] = false

			if activeColor == chess.White {
				changes[domain.GameWhiteTimeJsonTag] = 0
				changes[domain.GameResultJsonTag] = chess.BlackWon.String()
			} else {
				changes[domain.GameBlackTimeJsonTag] = 0
				changes[domain.GameResultJsonTag] = chess.WhiteWon.String()
			}

			updated, err := c.gameRepo.Update(ctx, gameID, version, changes)
			if err != nil {
				log.Printf("Usecase/Game/handleTimer, error updating: %v", err)
			}
			if updated && err == nil {
				c.timerManager.StopAndDeleteTimer(gameID)
				onTimeOut(changes)
			}
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
	onTimeOut func(utils.Changes[domain.GameFieldJsonTag]),
) (changes utils.Changes[domain.GameFieldJsonTag], updated bool, err error) {
	g, err := c.gameRepo.Get(ctx, gameID)
	if err != nil {
		return nil, false, err
	}

	changes, activeColor, err := makeMove(g, playerID, move)
	if err != nil {
		return nil, false, err
	}

	updated, err = c.gameRepo.Update(ctx, gameID, g.Version, changes)
	if err != nil {
		return nil, false, err
	}
	if !updated {
		return nil, false, nil
	}

	var timerDuration time.Duration
	if activeColor == chess.White {
		timerDuration = intToMillisecondsDuration(g.WhiteTime)
	} else {
		timerDuration = intToMillisecondsDuration(g.BlackTime)
	}

	_, gameOver := changes["Result"]

	c.handleTimer(
		context.Background(),
		onTimeOut,
		gameID,
		g.Version+1,
		timerDuration,
		activeColor,
		gameOver,
	)

	return changes, true, nil
}

func (c gameUseCase) UpdateDraw(
	ctx context.Context,
	gameID int,
	whiteDrawStatus bool,
	blackDrawStatus bool,
) (changes utils.Changes[domain.GameFieldJsonTag], updated bool, err error) {
	game, err := c.gameRepo.Get(ctx, gameID)
	if err != nil {
		return nil, false, err
	}

	if game.Result != "" {
		return nil, false, nil
	}

	changes = make(utils.Changes[domain.GameFieldJsonTag])
	changes[domain.GameWhiteDrawStatusJsonTag] = whiteDrawStatus
	changes[domain.GameBlackDrawStatusJsonTag] = blackDrawStatus

	updated, err = c.gameRepo.Update(ctx, gameID, game.Version, changes)
	if err != nil {
		return nil, false, err
	}

	return changes, updated, nil
}

func (c gameUseCase) UpdateResult(
	ctx context.Context,
	gameID int,
	method string,
	result string,
) (changes utils.Changes[domain.GameFieldJsonTag], updated bool, err error) {
	game, err := c.gameRepo.Get(ctx, gameID)
	if err != nil {
		return nil, false, err
	}

	if game.Result != "" {
		return nil, false, nil
	}

	changes = make(utils.Changes[domain.GameFieldJsonTag])
	changes[domain.GameMethodJsonTag] = method
	changes[domain.GameResultJsonTag] = result

	updated, err = c.gameRepo.Update(ctx, gameID, game.Version, changes)
	if err != nil {
		return nil, false, err
	}

	return changes, updated, nil
}
