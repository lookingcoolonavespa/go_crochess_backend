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
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
	"github.com/notnil/chess"
)

var timeNow = time.Now

type gameUseCase struct {
	db           *sql.DB
	gameRepo     domain.GameRepo
	timerManager *domain_timerManager.TimerManager
	gameCache    map[int]*chess.Game
}

func NewGameUseCase(
	db *sql.DB,
	gameRepo domain.GameRepo,
) gameUseCase {
	return gameUseCase{
		db,
		gameRepo,
		domain_timerManager.NewTimerManager(),
		make(map[int]*chess.Game),
	}
}

func getOnTimeOut(
	room domain.Room,
	gameID int,
) func(changes domain.GameChanges) {
	return func(changes domain.GameChanges) {
		jsonData, err := domain_websocket.NewOutboundMessage(
			fmt.Sprint(domain_websocket.GameTopic, "/", gameID),
			domain_websocket.TimeOutEvent,
			changes,
		).
			ToJSON("UseCase/Game/OnTimeOut, error converting data to json, err: %v\n")
		if err != nil {
			jsonData, err := domain_websocket.NewOutboundMessage(
				"error", domain_websocket.ErrorEvent,
				"game timer ran out, but there was an error converting the update to json",
			).ToJSON(
				"UseCase/Game/OnTimeOut, error converting data to json, err: %v\n",
			)
			log.Printf("UseCase/Game/OnTimeOut, game timer ran out, but there was an error converting the update to json\nerr: %v", err)

			room.BroadcastMessage(jsonData)
			return
		}

		room.BroadcastMessage(jsonData)
	}
}

func (c gameUseCase) OnAccept(
	ctx context.Context,
	g domain.Game,
	r domain.Room,
) (gameID int, err error) {
	g.TimeStampAtTurnStart = timeNow().UnixMilli()
	g.WhiteTime = g.Time
	g.BlackTime = g.Time

	gameID, err = c.gameRepo.Insert(ctx, g)
	if err != nil {
		return -1, err
	}

	if g.WhiteID != "engine" && g.BlackID != "engine" {
		c.handleTimer(
			ctx,
			getOnTimeOut(r, gameID),
			gameID,
			1,
			intToMillisecondsDuration(g.Time),
			chess.White,
			false,
		)
	}

	return gameID, nil
}

func (c gameUseCase) Get(ctx context.Context, gameID int) (domain.Game, error) {
	game, err := c.gameRepo.Get(ctx, gameID)
	if err != nil {
		return domain.Game{}, err
	}

	return game, nil
}

func (c gameUseCase) makeMove(
	g domain.Game,
	playerID string,
	move string,
) (domain.GameChanges, chess.Color, error) {
	// makeMove returns the changes that need to be made to game structured as key/value pairs,
	// the active color, and errors
	changes := make(domain.GameChanges)

	gameState, ok := c.gameCache[g.ID]
	if !ok {
		gameState = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
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

		c.gameCache[g.ID] = gameState
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

	changes[domain.GameMovesJsonTag] = move

	return changes, activeColor.Other(), nil
}

func (c gameUseCase) handleTimer(
	ctx context.Context,
	onTimeOut func(domain.GameChanges),
	gameID int, version int,
	duration time.Duration,
	activeColor chess.Color,
	gameOver bool,
) {
	if gameOver {
		c.timerManager.StopAndDeleteTimer(gameID)
	} else {
		c.timerManager.StartTimer(gameID, duration, func() {
			changes := make(domain.GameChanges)
			changes[domain.GameMethodJsonTag] = "TimeOut"
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
	room domain.Room,
) (changes domain.GameChanges, updated bool, err error) {
	g, err := c.gameRepo.Get(ctx, gameID)
	if err != nil {
		return nil, false, err
	}

	if g.Result != "" {
		return nil, false, nil
	}

	changes, activeColor, err := c.makeMove(g, playerID, move)
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

	_, gameOver := changes[domain.GameResultJsonTag]

	if g.WhiteID != "engine" && g.BlackID != "engine" {
		c.handleTimer(
			context.Background(),
			getOnTimeOut(room, gameID),
			gameID,
			g.Version+1,
			timerDuration,
			activeColor,
			gameOver,
		)
	}

	return changes, true, nil
}

func (c gameUseCase) UpdateDraw(
	ctx context.Context,
	gameID int,
	whiteDrawStatus bool,
	blackDrawStatus bool,
) (changes domain.GameChanges, updated bool, err error) {
	game, err := c.gameRepo.Get(ctx, gameID)
	if err != nil {
		return nil, false, err
	}

	if game.Result != "" {
		return nil, false, nil
	}

	changes = make(domain.GameChanges)
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
) (changes domain.GameChanges, updated bool, err error) {
	game, err := c.gameRepo.Get(ctx, gameID)
	if err != nil {
		return nil, false, err
	}

	if game.Result != "" {
		return nil, false, nil
	}

	changes = make(domain.GameChanges)
	changes[domain.GameMethodJsonTag] = method
	changes[domain.GameResultJsonTag] = result

	updated, err = c.gameRepo.Update(ctx, gameID, game.Version, changes)
	if err != nil {
		return nil, false, err
	}

	return changes, updated, nil
}
