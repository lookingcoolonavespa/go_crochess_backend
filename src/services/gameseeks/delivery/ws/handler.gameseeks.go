package delivery_ws_gameseeks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
)

const topicName = "gameseeks"

// events
const (
	AcceptEvent = "accept"
)

type GameseeksHandler struct {
	topic   domain_websocket.TopicWithoutParm
	usecase domain.GameseeksUseCase
	repo    domain.GameseeksRepo
}

type AcceptedGameseek struct {
	GameID      int
	PlayerColor domain.Color
}

func NewGameseeksHandler(
	repo domain.GameseeksRepo,
	usecase domain.GameseeksUseCase,
	topic domain_websocket.Topic,
) GameseeksHandler {
	handler := GameseeksHandler{
		topic.(domain_websocket.TopicWithoutParm),
		usecase,
		repo,
	}

	return handler
}

func (g GameseeksHandler) HandlerGetGameseeksList(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	_ []byte,
) error {
	room.RegisterClient(client)

	list, err := g.repo.List(ctx)
	if err != nil {
		log.Printf("%s : %v", "GameseeksHandler/HandlerGetGameseeksList/List/ShouldFindList", err)
		return errors.New(fmt.Sprintf("There was an error retreiving game seeks. %v", err))
	}

	err = client.SendMessage(
		topicName,
		domain_websocket.InitEvent,
		list,
		"GameseeksHandler/HandlerGetGameseeksList/List/ShouldEncodeIntoJson : %v",
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
			"GameseeksHandler/HandleGameseekInsert, Failed to unmarshal message: %v\n",
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

	whiteID, err := strconv.Atoi(game.WhiteID)
	if err != nil {
		log.Printf("unable to convert white id: %v to string", game.WhiteID)
		return err
	}
	whiteClient, ok := g.topic.GetClient(whiteID)
	if !ok {
		return errors.New(fmt.Sprintf(`client "%v" is not subscribed to %s`, whiteID, topicName))
	}

	blackID, err := strconv.Atoi(game.BlackID)
	if err != nil {
		log.Printf("unable to convert black id: %v to string", game.BlackID)
		return err
	}
	blackClient, ok := g.topic.GetClient(blackID)
	if !ok {
		return errors.New(fmt.Sprintf(`client "%v" is not subscribed to %s`, blackID, topicName))
	}

	gameID, deletedGameseeks, err := g.usecase.OnAccept(ctx, game)
	if err != nil {
		return err
	}

	game.ID = int(gameID)

	jsonDeletedGameseeks, err := domain_websocket.NewOutboundMessage(
		topicName,
		"deletion",
		deletedGameseeks).
		ToJSON("HandlerGameseeks/HandlerAcceptGameseek: error transforming message to json\nerr: %v")
	if err != nil {
		return err
	}

	jsonWhiteMessage, err := domain_websocket.NewOutboundMessage(
		fmt.Sprint("user/", topicName),
		"accepted",
		AcceptedGameseek{
			game.ID,
			domain.White,
		}).
		ToJSON("HandlerGameseeks/HandlerAcceptGameseek: error transforming message to json\nerr: %v")
	if err != nil {
		return err
	}

	jsonBlackMessage, err := domain_websocket.NewOutboundMessage(
		fmt.Sprint("user/", topicName),
		"accepted",
		AcceptedGameseek{
			game.ID,
			domain.Black,
		}).
		ToJSON("HandlerGameseeks/HandlerAcceptGameseek: error transforming message to json\nerr: %v")
	if err != nil {
		return err
	}

	room.BroadcastMessage(jsonDeletedGameseeks)
	whiteClient.SendBytes(jsonWhiteMessage)
	blackClient.SendBytes(jsonBlackMessage)

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
		"deletion",
		deletedGameseeks).
		ToJSON("HandlerGameseeks/HandlerOnUnsubscribe: error transforming message to json\nerr: %v")
	if err != nil {
		return err
	}

	room.BroadcastMessage(jsonDeletedGameseeks)

	return nil
}
