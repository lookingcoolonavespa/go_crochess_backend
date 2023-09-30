package delivery_ws_gameseeks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bxcodec/faker"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/repository/mock"
	mock_usecase_gameseeks "github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/usecase/mock"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
	domain_websocket_mock "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket/mock"
	"github.com/stretchr/testify/assert"
)

func TestGameseeksHandler_HandlerGetGameseeksList(t *testing.T) {
	var mockGameseek domain.Gameseek

	err := faker.FakeData(&mockGameseek)
	assert.NoError(t, err)

	mockRepo := new(repository_gameseeks_mock.GameseeksMockRepo)
	mockUseCase := new(mock_usecase_gameseeks.GameseeksMockUseCase)

	mockGameseeks := make([]domain.Gameseek, 0)
	mockGameseeks = append(mockGameseeks, mockGameseek)

	mockRepo.On("List", context.Background()).Return(mockGameseeks, nil).Once()

	topic, err := domain_websocket.NewTopic("topic")
	assert.NoError(t, err)
	r := NewGameseeksHandler(mockRepo, mockUseCase, topic)

	client := domain_websocket.NewClient(0, nil, nil)

	err = r.HandlerGetGameseeksList(context.Background(), nil, client, nil)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestGameseeksHandler_HandlerInsertGameseek(t *testing.T) {
	var mockGameseek domain.Gameseek

	err := faker.FakeData(&mockGameseek)
	assert.NoError(t, err)

	mockRepo := new(repository_gameseeks_mock.GameseeksMockRepo)
	mockUseCase := new(mock_usecase_gameseeks.GameseeksMockUseCase)

	mockRepo.On("Insert", context.Background(), mockGameseek).Return(nil).Once()

	topic, err := domain_websocket.NewTopic("topic")
	assert.NoError(t, err)

	r := NewGameseeksHandler(mockRepo, mockUseCase, topic)

	jsonData, err := json.Marshal(mockGameseek)
	assert.NoError(t, err)

	testChannel := make(chan []byte)
	client := domain_websocket_mock.NewMockClient(testChannel)
	room := domain_websocket.NewRoom([]domain_websocket.Client{client}, "")

	err = r.HandleGameseekInsert(context.Background(), room, nil, jsonData)

	receivedMessage := <-testChannel
	assert.Equal(t, jsonData, receivedMessage)

	mockRepo.AssertExpectations(t)
}
