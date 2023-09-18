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
}

func NewGameUseCase(gameseeksRepo domain.GameseeksRepo, gameRepo domain.GameRepo) gameUseCase {
	return gameUseCase{gameseeksRepo, gameRepo}
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

func updater(g *domain.Game, playerID string, move string, changes map[string]interface{}) error {
	gameState := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	moves := strings.Split(g.Moves, " ")

	for _, move := range moves {
		err := gameState.MoveStr(move)
		if err != nil {
			return err
		}
	}

	activeColor := gameState.Position().Turn().Name()
	if activeColor == "White" && g.WhiteID != playerID ||
		activeColor == "Black" && g.BlackID != playerID {
		return errors.New("Invalid player.")
	}

	err := gameState.MoveStr(move)
	if err != nil {
		return err
	}

	outcome := gameState.Outcome().String()
	if outcome != string(chess.NoOutcome) {
		changes["Result"] = g.Result
		changes["Method"] = g.Method

	} else {
		if elgibleDraw := len(gameState.EligibleDraws()) > 0; elgibleDraw {
			g.DrawRecord.Black = true
			g.DrawRecord.White = true
		}

		timeSpent := time.Now().Unix() - g.TimeStampAtTurnStart

		var activeTime int
		var fieldOfActiveTime string
		if activeColor == "White" {
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

	return nil
}

func (c gameUseCase) UpdateOnMove(gameID int, playerID string, move string) error {
	g, err := c.gameRepo.Get(gameID)
	if err != nil {
		return err
	}

	changes := make(map[string]interface{})
	err = updater(g, playerID, move, changes)
	if err != nil {
		return err
	}

	err = c.gameRepo.Update(gameID, changes)
	if err != nil {
		return err
	}

	return nil
}
