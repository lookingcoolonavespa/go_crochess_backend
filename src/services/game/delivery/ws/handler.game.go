package delivery_ws_game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/delivery_utils"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
)

const baseTopicName = "game"
const jsonErrorMessage = "Handler/Game/HandlerUpdateDraw, Failed to convert message to json: %v\n"

type GameHandler struct {
	usecase domain.GameUseCase
}

func NewGameHandler(
	usecase domain.GameUseCase,
) GameHandler {
	return GameHandler{
		usecase,
	}

}

func (g GameHandler) HandlerOnSubscribe(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	_ []byte,
) error {
	err := client.Subscribe(room)

	param, err := room.GetParam()
	if err != nil {
		log.Printf("Handler/Game/HandlerGetGame: room is missing param")
		return err
	}

	gameID, err := strconv.Atoi(param)
	if err != nil {
		log.Printf("Handler/Game/HandlerGetGame: ran into an error converting room.Param to int\nroom.Param: %v\nerr: %v", param, err)
		return err
	}

	game, err := g.usecase.Get(ctx, gameID)
	if err != nil {
		log.Printf("Handler/Game/HandlerGetGame: ran into an error getting game\nerr: %v", err)
		return err
	}

	err = client.SendMessage(
		fmt.Sprintf("%s/%s", baseTopicName, param),
		domain_websocket.InitEvent,
		game,
		"Handler/Game/HandlerGetGame: error turning game into json\nerr: %v",
	)

	return err
}

func (g GameHandler) HandlerOnUnsubscribe(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	_ []byte,
) error {
	client.Unsubscribe(room)
	return nil
}

func (g GameHandler) HandlerMakeMove(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	payload []byte,
) error {
	gID, err := room.GetParam()
	if err != nil {
		log.Printf("Handler/Game/HandlerMakeMove: room is missing param")
		return err
	}

	gameID, err := strconv.Atoi(gID)
	if err != nil {
		log.Printf("Handler/Game/HandlerMakeMove: param is not a valid int")
		return err
	}

	type MovePayload struct {
		PlayerID string `json:"player_id"`
		Move     string `json:"move"`
	}
	var movePayload MovePayload
	err = json.Unmarshal(payload, &movePayload)
	if err != nil {
		log.Printf("Handler/Game/HandlerMakeMove: failed to unmarshal payload, err: %v\n", err)
		return err
	}

	missingFields := make([]string, 0)
	if movePayload.PlayerID == "" {
		missingFields = append(missingFields, "player_id")
	}
	if movePayload.Move == "" {
		missingFields = append(missingFields, "move")
	}
	if len(missingFields) != 0 {
		errorMessage := fmt.Sprintf("move is missing the following fields: %v", strings.Join(missingFields, ", "))
		err = client.SendError(
			errorMessage,
			"Handler/Game/HandlerMakeMove, Failed to convert message to json: %v\n",
		)
		if err != nil {
			return err
		}

		return errors.New(errorMessage)
	}

	changes, updated, err := g.usecase.UpdateOnMove(
		ctx,
		gameID,
		movePayload.PlayerID,
		movePayload.Move,
		delivery_utils.GetOnTimeOut(room, client, &gameID, jsonErrorMessage),
	)
	if err != nil {
		return err
	}
	if !updated {
		client.SendError(
			`Unable to make move because either the game is over or
            because the game was updated before your request could be completed`,
			jsonErrorMessage,
		)
		return nil
	}

	var event string
	if changes[domain.GameResultJsonTag] == nil {
		event = domain_websocket.MakeMoveEvent
	} else {
		event = domain_websocket.GameOverEvent
	}

	jsonData, err := domain_websocket.NewOutboundMessage(
		fmt.Sprint(baseTopicName, "/", gameID),
		event,
		changes,
	).
		ToJSON(jsonErrorMessage)
	if err != nil {
		return err
	}

	room.BroadcastMessage(jsonData)

	return nil
}

func (g GameHandler) HandlerUpdateDraw(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	payload []byte,
) error {
	gID, err := room.GetParam()
	if err != nil {
		log.Printf("Handler/Game/HandlerUpdateDraw: room is missing param")
		return err
	}

	gameID, err := strconv.Atoi(gID)
	if err != nil {
		log.Printf("Handler/Game/HandlerUpdateDraw: param is not a valid int")
		return err
	}

	type UpdateDrawPayload struct {
		White bool `json:"white"`
		Black bool `json:"black"`
	}
	var updateDrawPayload UpdateDrawPayload
	err = json.Unmarshal(payload, &updateDrawPayload)
	if err != nil {
		log.Printf("Handler/Game/HandlerUpdateDraw: failed to unmarshal payload, err: %v\n", err)
		return err
	}

	changes, updated, err := g.usecase.UpdateDraw(ctx, gameID, updateDrawPayload.White, updateDrawPayload.Black)
	if err != nil {
		return err
	}
	if !updated {
		client.SendError(
			`Unable to update draw status because either the game is over or
            because the game was updated before your request could be completed`,
			jsonErrorMessage,
		)
		return nil
	}

	jsonData, err := domain_websocket.NewOutboundMessage(
		fmt.Sprint(baseTopicName, "/", gameID),
		domain_websocket.UpdateDrawEvent,
		changes,
	).
		ToJSON(jsonErrorMessage)
	if err != nil {
		return err
	}

	room.BroadcastMessage(jsonData)

	return nil
}

func (g GameHandler) HandlerUpdateResult(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	payload []byte,
) error {
	gID, err := room.GetParam()
	if err != nil {
		log.Printf("Handler/Game/HandlerUpdateResult: room is missing param")
		return err
	}

	gameID, err := strconv.Atoi(gID)
	if err != nil {
		log.Printf("Handler/Game/HandlerUpdateResult: param is not a valid int")
		return err
	}

	type UpdateResultPayload struct {
		Method string `json:"method"`
		Result string `json:"result"`
	}
	var updateResultPayload UpdateResultPayload
	err = json.Unmarshal(payload, &updateResultPayload)
	if err != nil {
		log.Printf("Handler/Game/HandlerUpdateResult: failed to unmarshal payload, err: %v\n", err)
		return err
	}

	changes, updated, err := g.usecase.UpdateResult(
		ctx,
		gameID,
		updateResultPayload.Method,
		updateResultPayload.Result,
	)
	if err != nil {
		return err
	}
	if !updated {
		client.SendError(
			`Unable to update result because either the game is over or
            because the game was updated before your request could be completed`,
			jsonErrorMessage,
		)
		return nil
	}

	jsonData, err := domain_websocket.NewOutboundMessage(
		fmt.Sprint(baseTopicName, "/", gameID),
		domain_websocket.UpdateResultEvent,
		changes,
	).
		ToJSON(jsonErrorMessage)
	if err != nil {
		return err
	}

	room.BroadcastMessage(jsonData)

	return nil
}
