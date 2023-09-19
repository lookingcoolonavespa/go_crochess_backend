package usecase_game

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lookingcoolonavespa/go_crochess_backend/domain"
	"github.com/notnil/chess"
)

type gameUseCase struct {
	gameseeksRepo domain.GameseeksRepo
	gameRepo      domain.GameRepo
	timerManager  *domain.TimerManager
}

func NewGameUseCase(gameseeksRepo domain.GameseeksRepo, gameRepo domain.GameRepo, timerManager *domain.TimerManager) gameUseCase {
	return gameUseCase{gameseeksRepo, gameRepo, timerManager}
}

func (c gameUseCase) Insert(g *domain.Game) error {
	err := c.gameRepo.Insert(g)
	if err != nil {
		return err
	}
	err = c.gameseeksRepo.Delete(g.WhiteID, g.BlackID)
	if err != nil {
		return err
	}

	return nil
}

func makeMove(g *domain.Game, playerID string, move string) (map[string]interface{}, chess.Color, error) {
	// makeMove returns the changes that need to be made to game structured as key/value pairs,
	// the active color, and errors
	changes := make(map[string]interface{})

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

	return changes, activeColor, nil
}

func (c gameUseCase) handleTimer(gameID int, version int, duration time.Duration, activeColor chess.Color, onGameOver func()) {
	c.timerManager.StopAndDeleteTimer(fmt.Sprint(gameID))

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
		c.gameRepo.Update(gameID, version, changes)
		onGameOver()
	})
}

func intToMillisecondsDuration(value int) time.Duration {
	return time.Duration(value) * time.Millisecond
}

func (c gameUseCase) UpdateOnMove(
	gameID int,
	playerID string,
	move string,
	onGameOver func(),
) error {
	g, err := c.gameRepo.Get(gameID)
	if err != nil {
		return err
	}

	changes, activeColor, err := makeMove(g, playerID, move)
	if err != nil {
		return err
	}

	var timerDuration time.Duration
	if activeColor == chess.White {
		timerDuration = intToMillisecondsDuration(g.WhiteTime)
	} else {
		timerDuration = intToMillisecondsDuration(g.BlackTime)
	}

	updated, err := c.gameRepo.Update(gameID, g.Version, changes)
	if !updated {
		return errors.New("The move did not reach the server fast enough")
	}
	if err != nil {
		return err
	}

	c.handleTimer(gameID, g.Version+1, timerDuration, activeColor, onGameOver)

	return nil
}
