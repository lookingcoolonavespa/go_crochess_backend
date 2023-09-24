package delivery_ws_gameseeks

import (
	"encoding/json"
	"testing"

	"github.com/bxcodec/faker"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/repository/mock"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGameseeksHandler_HandlerGetGameseeksList(t *testing.T) {
	var mockGameseek domain.Gameseek

	err := faker.FakeData(&mockGameseek)
	assert.NoError(t, err)

	mockRepo := new(repository_gameseeks_mock.GameseeksMockRepo)

	mockGameseeks := make([]domain.Gameseek, 0)
	mockGameseeks = append(mockGameseeks, mockGameseek)

	mockRepo.On("List").Return(mockGameseeks, nil).Once()

	topic, err := domain_websocket.NewTopic("topic")
	assert.NoError(t, err)
	r := NewGameseeksHandler(mockRepo, *topic)

	client := domain_websocket.NewClient(nil, nil)

	err = r.HandlerGetGameseeksList(client, nil)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestGameseeksHandler_HandlerInsertGameseek(t *testing.T) {
	var mockGameseek domain.Gameseek

	err := faker.FakeData(&mockGameseek)
	assert.NoError(t, err)

	mockRepo := new(repository_gameseeks_mock.GameseeksMockRepo)

	mockRepo.On("Insert", mock.AnythingOfType("*domain.Gameseek")).Return(nil).Once()

	topic, err := domain_websocket.NewTopic("topic")
	assert.NoError(t, err)

	r := NewGameseeksHandler(mockRepo, *topic)

	jsonData, err := json.Marshal(mockGameseek)
	assert.NoError(t, err)

	err = r.HandleGameseekInsert(nil, jsonData)

	mockRepo.AssertExpectations(t)
}
