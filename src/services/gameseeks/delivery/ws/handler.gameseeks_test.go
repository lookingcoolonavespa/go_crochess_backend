package delivery_ws_gameseeks

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/bxcodec/faker"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/repository/mock"
	mock_usecase_gameseeks "github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/usecase/mock"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
	"github.com/stretchr/testify/assert"
)

func TestGameseeksHandler_HandlerOnSubscribe(t *testing.T) {
	var mockGameseek domain.Gameseek

	err := faker.FakeData(&mockGameseek)
	assert.NoError(t, err)

	mockRepo := new(repository_gameseeks_mock.GameseeksMockRepo)
	mockUseCase := new(mock_usecase_gameseeks.GameseeksMockUseCase)

	mockGameseeks := make([]domain.Gameseek, 0)
	mockGameseeks = append(mockGameseeks, mockGameseek)

	mockRepo.On("List", context.Background()).Return(mockGameseeks, nil).Once()

	r := NewGameseeksHandler(mockRepo, mockUseCase, domain_websocket.TopicWithParam{})

	messageChan := make(chan []byte)
	client := domain_websocket.NewClient("0", messageChan, nil, nil)

	room := domain_websocket.NewRoom(make([]*domain_websocket.Client, 0), "")
	err = r.HandlerOnSubscribe(context.Background(), room, client, nil)
	assert.NoError(t, err)

	select {
	case message := <-messageChan:
		assert.Contains(t, string(message), domain_websocket.InitEvent)

	case <-time.After(1 * time.Second):
		t.Fatal("TestGameseeksHandler_HandlerGetGameseeksList hanging waiting for message")
	}

	_, subscribed := room.GetClient(client.GetID())
	assert.True(t, subscribed)

	mockRepo.AssertExpectations(t)
}

func TestGameseeksHandler_HandlerInsertGameseek(t *testing.T) {
	var mockGameseek domain.Gameseek

	err := faker.FakeData(&mockGameseek)
	assert.NoError(t, err)

	mockRepo := new(repository_gameseeks_mock.GameseeksMockRepo)
	mockUseCase := new(mock_usecase_gameseeks.GameseeksMockUseCase)

	mockRepo.On("Insert", context.Background(), mockGameseek).Return(nil).Once()

	r := NewGameseeksHandler(mockRepo, mockUseCase, domain_websocket.TopicWithParam{})

	jsonData, err := json.Marshal(mockGameseek)
	assert.NoError(t, err)

	testChannel := make(chan []byte)
	client := domain_websocket.NewClient("0", testChannel, nil, nil)
	room := domain_websocket.NewRoom([]*domain_websocket.Client{client}, "")

	err = r.HandleGameseekInsert(context.Background(), room, nil, jsonData)

	receivedMessage := <-testChannel
	assert.Contains(t, string(receivedMessage), string(jsonData))

	mockRepo.AssertExpectations(t)
}

func TestGameseeksHandler_HandlerOnUnsubscribe(t *testing.T) {
	mockRepo := new(repository_gameseeks_mock.GameseeksMockRepo)
	mockUseCase := new(mock_usecase_gameseeks.GameseeksMockUseCase)

	unsubscribedChannel := make(chan []byte)
	unsubscribeClient := domain_websocket.NewClient("0", unsubscribedChannel, nil, nil)

	deletedGameseeks := []int{1, 2}
	mockRepo.On("DeleteFromSeeker", context.Background(), unsubscribeClient.GetID()).
		Return(deletedGameseeks, nil).
		Once()

	r := NewGameseeksHandler(mockRepo, mockUseCase, domain_websocket.TopicWithParam{})

	subscribedChannel := make(chan []byte)
	subscribedClient := domain_websocket.NewClient("0", subscribedChannel, nil, nil)
	room := domain_websocket.NewRoom(
		[]*domain_websocket.Client{unsubscribeClient, subscribedClient},
		"",
	)

	err := r.HandlerOnUnsubscribe(context.Background(), room, unsubscribeClient, []byte{})
	assert.NoError(t, err)

	timeOutChan := make(chan string)
	time.AfterFunc(time.Second*2, func() { timeOutChan <- "time out" })
	select {
	case receivedMessage := <-subscribedChannel:
		assert.Contains(t, string(receivedMessage), string(domain_websocket.DeletionEvent))
	case <-unsubscribedChannel:
		assert.Fail(t, "unsubscribed channel received message")
	case <-timeOutChan:
		break
	}

	_, subscribed := room.GetClient(unsubscribeClient.GetID())
	assert.False(t, subscribed)

	mockRepo.AssertExpectations(t)
}
