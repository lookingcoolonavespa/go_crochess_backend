package delivery_ws_gameseeks

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
)

type GameseeksHandler struct {
	repo domain.GameseeksRepo
}

func NewGameseeksHandler(repo domain.GameseeksRepo, topic domain_websocket.Topic) *GameseeksHandler {
	handler := &GameseeksHandler{repo}

	topic.RegisterEvent(domain_websocket.SubscribeEvent, handler.HandlerGetGameseeksList)
	topic.RegisterEvent(domain_websocket.InsertEvent, handler.HandleGameseekInsert)

	return handler
}

func (g *GameseeksHandler) HandlerGetGameseeksList(client *domain_websocket.Client, _ []byte) error {
	list, err := g.repo.List()
	if err != nil {
		log.Printf("%s : %v", "GameseeksHandler/HandlerGetGameseeksList/List/ShouldFindList", err)
		return errors.New(fmt.Sprintf("There was an error retreiving game seeks. %v", err))
	}

	jsonData, err := json.Marshal(list)
	if err != nil {
		log.Printf("%s : %v", "GameseeksHandler/HandlerGetGameseeksList/List/ShouldEncodeIntoJson", err)
		return errors.New(fmt.Sprintf("There was an error retreiving game seeks. %v", err))
	}

	client.Send <- []byte(jsonData)
	return nil
}

func (g *GameseeksHandler) HandleGameseekInsert(_ *domain_websocket.Client, jsonGameseek []byte) error {
	var param domain.Gameseek
	if err := json.Unmarshal(jsonGameseek, &param); err != nil {
		return errors.New(fmt.Sprintf("Failed to decode request body: %v", err))
	}

	err := g.repo.Insert(&param)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to save game seek: %v", err))
	}

	return nil
}
