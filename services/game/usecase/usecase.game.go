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

func (c gameUseCase) Insert(g *domain.Game) {
	c.gameRepo.Insert(g)
	c.gameseeksRepo.Delete(g.BlackID, g.WhiteID)
}

func (c gameUseCase) UpdateOnMove(gameID int, playerID string, move string) error {
	updater := func(g *domain.Game, changes map[string]interface{}) error {
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

			var activeTime int64
			var fieldOfActiveTime string
			if activeColor == "White" {
				activeTime = g.WhiteTime
				fieldOfActiveTime = "WhiteTime"
			} else {
				activeTime = g.BlackTime
				fieldOfActiveTime = "BlackTime"
			}

			base := activeTime - timeSpent
			changes[fieldOfActiveTime] = base + (int64(g.Increment * 1000))
			changes["TimeStampAtTurnStart"] = time.Now().Unix()
		}

		changes["Moves"] = g.Moves + fmt.Sprintf("%s ", move)

		gameState.ChangeNotation(chess.AlgebraicNotation{})
		changes["History"] = g.History

		return nil
	}

	err := c.gameRepo.Update(gameID, updater)
	if err != nil {
		return err
	}

	return nil
}
