package delivery_ws_game

import (
	"context"
	"fmt"
	"log"
	"strconv"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
)

const baseTopicName = "game"

type GameHandler struct {
	topic   domain_websocket.TopicWithParam
	usecase domain.GameUseCase
}

func NewGameHandler(
	topic domain_websocket.TopicWithParam,
	usecase domain.GameUseCase,
) GameHandler {
	return GameHandler{
		topic,
		usecase,
	}

}

func (g GameHandler) HandlerOnSubscribe(
	ctx context.Context,
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	_ []byte,
) error {
	client.Subscribe(room)

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
