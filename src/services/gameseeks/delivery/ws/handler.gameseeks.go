package delivery_ws_gameseeks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

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

	topic.RegisterEvent(domain_websocket.SubscribeEvent, handler.HandlerGetGameseeksList)
	topic.RegisterEvent(domain_websocket.InsertEvent, handler.HandleGameseekInsert)

	return handler
}

func (g GameseeksHandler) HandlerGetGameseeksList(ctx context.Context, _ domain_websocket.Room, client domain_websocket.Client, _ []byte) error {
	list, err := g.repo.List(ctx)
	if err != nil {
		log.Printf("%s : %v", "GameseeksHandler/HandlerGetGameseeksList/List/ShouldFindList", err)
		return errors.New(fmt.Sprintf("There was an error retreiving game seeks. %v", err))
	}

	jsonData, err := json.Marshal(list)
	if err != nil {
		log.Printf("%s : %v", "GameseeksHandler/HandlerGetGameseeksList/List/ShouldEncodeIntoJson", err)
		return errors.New(fmt.Sprintf("There was an error retreiving game seeks. %v", err))
	}

	client.Send(jsonData)
	return nil
}

func (g GameseeksHandler) HandleGameseekInsert(ctx context.Context, room domain_websocket.Room, _ domain_websocket.Client, jsonGameseek []byte) error {
	var gs domain.Gameseek
	if err := json.Unmarshal(jsonGameseek, &gs); err != nil {
		return errors.New(fmt.Sprintf("Failed to decode request body: %v", err))
	}

	err := g.repo.Insert(ctx, gs)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to save game seek: %v", err))
	}

	room.BroadcastMessage(jsonGameseek)

	return nil
}

func (g GameseeksHandler) HandlerAcceptGameseek(
	ctx context.Context,
	room domain_websocket.Room,
	client domain_websocket.Client,
	payload []byte,
) error {
	var game domain.Game
	err := json.Unmarshal(payload, &game)
	if err != nil {
		return err
	}

	gameID, deletedGameseeks, err := g.usecase.OnAccept(ctx, game)
	if err != nil {
		return err
	}

	game.ID = int(gameID)

	whiteID, err := strconv.Atoi(game.WhiteID)
	if err != nil {
		return err
	}
	whiteClient, ok := g.topic.GetClient(whiteID)
	if !ok {
		return errors.New("")
	}

	blackID, err := strconv.Atoi(game.BlackID)
	if err != nil {
		return err
	}
	blackClient, ok := g.topic.GetClient(blackID)
	if !ok {
		return errors.New("")
	}

	jsonDeletedGameseeks, err := domain_websocket.NewOutboundMessage(
		topicName,
		"deletion",
		deletedGameseeks).
		ToJSON()
	if err != nil {
		return err
	}

	jsonDataWhite, err := domain_websocket.NewOutboundMessage(
		topicName,
		"accepted",
		AcceptedGameseek{
			game.ID,
			domain.White,
		}).
		ToJSON()
	if err != nil {
		return err
	}

	jsonDataBlack, err := domain_websocket.NewOutboundMessage(
		topicName,
		"accepted",
		AcceptedGameseek{
			game.ID,
			domain.Black,
		}).
		ToJSON()
	if err != nil {
		return err
	}

	whiteClient.Send(jsonDataWhite)
	blackClient.Send(jsonDataBlack)

	room.BroadcastMessage(jsonDeletedGameseeks)

	return nil
}
