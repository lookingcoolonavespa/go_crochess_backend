package delivery_utils

import (
	"fmt"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
)

func GetOnTimeOut(
	room *domain_websocket.Room,
	client *domain_websocket.Client,
	gameID *int,
	jsonErrorMessage string,
) func(changes domain.GameChanges) {
	return func(changes domain.GameChanges) {
		jsonData, err := domain_websocket.NewOutboundMessage(
			fmt.Sprint(domain_websocket.GameTopic, "/", *gameID),
			domain_websocket.TimeOutEvent,
			changes,
		).
			ToJSON(jsonErrorMessage)
		if err != nil {
			client.SendError(
				"game timer ran out, but there was an error converting the update to json",
				jsonErrorMessage,
			)
			return
		}

		room.BroadcastMessage(jsonData)
	}
}
