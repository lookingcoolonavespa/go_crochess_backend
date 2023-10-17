package delivery_ws_game

import (
	"context"
	"strconv"
	"testing"

	"github.com/bxcodec/faker"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	mock_usecase_game "github.com/lookingcoolonavespa/go_crochess_backend/src/services/game/usecase/mock"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
	"github.com/stretchr/testify/assert"
)

func TestGameHandler_HandlerOnSubscribe(t *testing.T) {
	var mockGame domain.Game

	err := faker.FakeData(&mockGame)
	assert.NoError(t, err)

	gameID := 516

	mockUseCase := new(mock_usecase_game.MockGameUseCase)

	mockUseCase.On("Get", context.Background(), gameID).Return(mockGame, nil).Once()

	gameIDStr := strconv.Itoa(gameID)

	h := NewGameHandler(mockUseCase)

	testChan := make(chan []byte)
	client := domain_websocket.NewClient("0", testChan, nil, nil)

	room := domain_websocket.NewRoom([]domain.Client{}, gameIDStr)

	err = h.HandlerOnSubscribe(context.Background(), room, client, make([]byte, 0))
	assert.NoError(t, err)

	select {
	case message := <-testChan:
		assert.Contains(t, string(message), gameIDStr)
		assert.Contains(t, string(message), domain_websocket.InitEvent)
	}

	_, subscribed := room.GetClient(client.GetID())
	assert.True(t, subscribed)
}

func TestGameHandler_HandlerOnUnsubscribe(t *testing.T) {
	gameID := 516

	mockUseCase := new(mock_usecase_game.MockGameUseCase)

	gameIDStr := strconv.Itoa(gameID)

	h := NewGameHandler(mockUseCase)

	testChan := make(chan []byte)
	client := domain_websocket.NewClient("0", testChan, nil, nil)

	room := domain_websocket.NewRoom([]domain.Client{client}, gameIDStr)

	err := h.HandlerOnUnsubscribe(context.Background(), room, client, make([]byte, 0))
	assert.NoError(t, err)

	_, subscribed := room.GetClient(client.GetID())
	assert.False(t, subscribed)
}
