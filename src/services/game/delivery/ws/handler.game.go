package delivery_ws_game

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
)

type GameHandler struct {
	topic   domain_websocket.TopicWithParm
	usecase domain.GameUseCase
}

func NewGameHandler(
	topic domain_websocket.TopicWithParm,
	usecase domain.GameUseCase,
) GameHandler {
	handler := GameHandler{
		topic,
		usecase,
	}

	topic.RegisterEvent(domain_websocket.SubscribeEvent, handler.HandlerGetGame)

	return handler
}

func (g GameHandler) HandlerGetGame(
	ctx context.Context,
	room domain_websocket.Room,
	client domain_websocket.Client,
	_ []byte,
) error {
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

	game, err := g.usecase.Get(context.Background(), gameID)
	if err != nil {
		log.Printf("Handler/Game/HandlerGetGame: ran into an error getting game, err: %v", err)
		return err
	}

	jsonData, err := json.Marshal(game)
	if err != nil {
		return err
	}

	client.Send(jsonData)

	return nil
}
