package delivery_utils

import (
	"fmt"
	"log"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
)

func GetOnTimeOut(
	room *domain_websocket.Room,
	gameID *int,
) func(changes domain.GameChanges) {
	return func(changes domain.GameChanges) {
		jsonData, err := domain_websocket.NewOutboundMessage(
			fmt.Sprint(domain_websocket.GameTopic, "/", gameID),
			domain_websocket.TimeOutEvent,
			changes,
		).
			ToJSON("UseCase/Game/OnTimeOut, error converting data to json, err: %v\n")
		if err != nil {
			jsonData, err := domain_websocket.NewOutboundMessage(
				"error", domain_websocket.ErrorEvent,
				"game timer ran out, but there was an error converting the update to json",
			).ToJSON(
				"UseCase/Game/OnTimeOut, error converting data to json, err: %v\n",
			)
			log.Printf("UseCase/Game/OnTimeOut, game timer ran out, but there was an error converting the update to json\nerr: %v", err)

			room.BroadcastMessage(jsonData)
			return
		}

		room.BroadcastMessage(jsonData)
	}
}
