package delivery_ws_gameseeks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/delivery_utils"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
)

const topicName = domain_websocket.GameseeksTopic

type GameseeksHandler struct {
	usecase   domain.GameseeksUseCase
	repo      domain.GameseeksRepo
	gameTopic domain_websocket.TopicWithParam
}

type AcceptedGameseek struct {
	GameID      int          `json:"game_id"`
	PlayerColor domain.Color `json:"playerColor"`
}

func NewGameseeksHandler(
	repo domain.GameseeksRepo,
	usecase domain.GameseeksUseCase,
	gameTopic domain_websocket.TopicWithParam,
) GameseeksHandler {
	handler := GameseeksHandler{
		usecase,
		repo,
		gameTopic,
	}

	return handler
}

func (g GameseeksHandler) HandlerOnSubscribe(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	_ []byte,
) error {
	err := client.Subscribe(room)
	if err != nil {
		return err
	}

	list, err := g.repo.List(ctx)
	if err != nil {
		log.Printf("%s : %v", "Handler/Gameseeks/HandlerGetGameseeksList/List/ShouldFindList", err)
		return errors.New(fmt.Sprintf("There was an error retreiving game seeks. %v", err))
	}

	err = client.SendMessage(
		topicName,
		domain_websocket.InitEvent,
		list,
		"Handler/Gameseeks/HandlerGetGameseeksList/List/ShouldEncodeIntoJson : %v",
	)
	return err
}

func (g GameseeksHandler) HandleGameseekInsert(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	jsonGameseek []byte,
) error {
	var gs domain.Gameseek
	if err := json.Unmarshal(jsonGameseek, &gs); err != nil {
		log.Printf("GameseeksHandler/HandleGameseekInsert, Failed to unmarshal message: %v\n", err)
		return err
	}

	filled, missingFields := gs.IsFilled()
	if !filled {
		errorMessage := fmt.Sprintf("gameseek is missing the following fields: %v", strings.Join(missingFields, ", "))
		err := client.SendError(
			errorMessage,
			"GameseeksHandler/HandleGameseekInsert, Failed to convert message to json: %v\n",
		)
		if err != nil {
			return err
		}

		return errors.New(errorMessage)
	}

	err := g.repo.Insert(ctx, gs)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to save gameseek: %v", err))
	}

	jsonMessage, err := domain_websocket.NewOutboundMessage(
		topicName,
		domain_websocket.InsertEvent,
		gs,
	).ToJSON("GameseeksHandler/HandleGameInsert, error converting message to json, err: %v")
	if err != nil {
		return err
	}

	room.BroadcastMessage(jsonMessage)

	return nil
}

func (g GameseeksHandler) HandlerAcceptGameseek(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	payload []byte,
) error {
	var game domain.Game
	err := json.Unmarshal(payload, &game)
	if err != nil {
		return err
	}

	filled, missingFields := game.IsFilledForInsert()
	if !filled {
		errorMessage := fmt.Sprintf("game missing the following fields: %v", strings.Join(missingFields, ", "))
		err := client.SendError(
			errorMessage,
			"GameseeksHandler/HandlerAcceptGameseek, Failed to unmarshal message: %v\n",
		)
		if err != nil {
			return err
		}

		return errors.New(errorMessage)
	}

	whiteClient, ok := room.GetClient(game.WhiteID)
	if !ok {
		log.Printf(
			`Handler/Gameseeks/HandlerAcceptGameseek, client "%v" is not subscribed to %s`,
			game.WhiteID,
			topicName,
		)
		return errors.New(fmt.Sprintf(`client "%v" is not subscribed to %s`, game.WhiteID, topicName))
	}

	blackClient, ok := room.GetClient(game.BlackID)
	if !ok {
		log.Printf(
			`Handler/Gameseeks/HandlerAcceptGameseek, client "%v" is not subscribed to %s`,
			game.BlackID,
			topicName,
		)
		return errors.New(fmt.Sprintf(`client "%v" is not subscribed to %s`, game.BlackID, topicName))
	}

	var gID int
	gameRoom := domain_websocket.NewRoom([]*domain_websocket.Client{}, "")
	gameID, err := g.usecase.OnAccept(
		ctx,
		game,
		delivery_utils.GetOnTimeOut(
			gameRoom,
			&gID,
		))
	if err != nil {
		return err
	}

	game.ID = gameID
	gID = gameID
	gameRoom.ChangeParam(fmt.Sprint(gameID))

	err = g.gameTopic.PushNewRoom(gameRoom)
	if err != nil {
		log.Printf("Handler/Gameseeks/HandlerAcceptGameseek, param of game room is an empty string")
		return err
	}

	jsonWhiteMessage, err := domain_websocket.NewOutboundMessage(
		topicName,
		domain_websocket.AcceptEvent,
		AcceptedGameseek{
			game.ID,
			domain.White,
		}).
		ToJSON("Handler/Gameseeks/HandlerAcceptGameseek: error transforming message to json\nerr: %v\n")
	if err != nil {
		return err
	}

	jsonBlackMessage, err := domain_websocket.NewOutboundMessage(
		topicName,
		domain_websocket.AcceptEvent,
		AcceptedGameseek{
			game.ID,
			domain.Black,
		}).
		ToJSON("HandlerGameseeks/HandlerAcceptGameseek: error transforming message to json\nerr: %v")
	if err != nil {
		return err
	}

	whiteClient.SendBytes(jsonWhiteMessage)
	blackClient.SendBytes(jsonBlackMessage)

	return nil
}

func (g GameseeksHandler) HandlerStartEngineGame(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	payload []byte,
) error {
	var game domain.Game
	err := json.Unmarshal(payload, &game)
	if err != nil {
		return err
	}

	filled, missingFields := game.IsFilledForInsert()
	if !filled {
		errorMessage := fmt.Sprintf("game missing the following fields: %v", strings.Join(missingFields, ", "))
		err := client.SendError(
			errorMessage,
			"Handler/Gameseeks/HandlerStartEngineGame, Failed to unmarshal message: %v\n",
		)
		if err != nil {
			return err
		}

		return errors.New(errorMessage)
	}

	var gID int
	gameRoom := domain_websocket.NewRoom([]*domain_websocket.Client{}, "")
	gameID, err := g.usecase.OnAccept(
		ctx,
		game,
		delivery_utils.GetOnTimeOut(
			gameRoom,
			&gID,
		))
	if err != nil {
		return err
	}

	game.ID = gameID
	gameRoom.ChangeParam(fmt.Sprint(gameID))

	err = g.gameTopic.PushNewRoom(gameRoom)
	if err != nil {
		log.Printf("Handler/Gameseeks/HandlerStartEngineGame, param of game room is an empty string")
		return err
	}

	var color domain.Color
	if game.WhiteID == client.GetID() {
		color = domain.White
	} else {
		color = domain.Black
	}

	jsonMessage, err := domain_websocket.NewOutboundMessage(
		topicName,
		domain_websocket.AcceptEvent,
		AcceptedGameseek{
			game.ID,
			color,
		}).
		ToJSON("Handler/Gameseeks/HandlerStartEngineGame: error transforming message to json\nerr: %v\n")
	if err != nil {
		return err
	}

	client.SendBytes(jsonMessage)

	return nil
}

func (g GameseeksHandler) HandlerOnUnsubscribe(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	_ []byte,
) error {
	client.Unsubscribe(room)

	deletedGameseeks, err := g.repo.DeleteFromSeeker(ctx, client.GetID())
	if err != nil {
		return err
	}

	jsonDeletedGameseeks, err := domain_websocket.NewOutboundMessage(
		topicName,
		domain_websocket.DeletionEvent,
		deletedGameseeks).
		ToJSON("HandlerGameseeks/HandlerOnUnsubscribe: error transforming message to json\nerr: %v")
	if err != nil {
		return err
	}

	room.BroadcastMessage(jsonDeletedGameseeks)

	return nil
}
