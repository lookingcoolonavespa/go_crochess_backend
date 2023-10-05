package delivery_ws_game

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/bxcodec/faker"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	mock_usecase_game "github.com/lookingcoolonavespa/go_crochess_backend/src/services/game/usecase/mock"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
	"github.com/stretchr/testify/assert"
)

func TestGameHandler_HandlerGetGame(t *testing.T) {
	var mockGame domain.Game

	err := faker.FakeData(&mockGame)
	assert.NoError(t, err)

	gameID := 516

	mockUseCase := new(mock_usecase_game.MockGameUseCase)

	mockUseCase.On("Get", context.Background(), gameID).Return(mockGame, nil).Once()

	gameIDStr := strconv.Itoa(gameID)

	topic, err := domain_websocket.NewTopic(fmt.Sprintf("game/%s", gameIDStr))
	assert.NoError(t, err)

	h := NewGameHandler(topic.(domain_websocket.TopicWithParam), mockUseCase)

	testChan := make(chan []byte)
	mockClient := domain_websocket.NewClient("0", testChan, nil, nil)

	mockRoom := domain_websocket.NewRoom([]*domain_websocket.Client{mockClient}, gameIDStr)

	err = h.HandlerOnSubscribe(context.Background(), mockRoom, mockClient, make([]byte, 0))
	assert.NoError(t, err)

	select {
	case message := <-testChan:
		assert.Contains(t, string(message), gameIDStr)
		assert.Contains(t, string(message), domain_websocket.InitEvent)
	}
}
