package delivery_http_gameseeks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/julienschmidt/httprouter"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
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

	r := httprouter.New()
	r = NewGameseeksHandler(r, mockRepo)

	req, err := http.NewRequest(http.MethodGet, "/gameseeks", nil)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
	mockRepo.AssertExpectations(t)
}

func TestGameseeksHandler_HandlerInsertGameseek(t *testing.T) {
	var mockGameseek domain.Gameseek

	err := faker.FakeData(&mockGameseek)
	assert.NoError(t, err)

	mockRepo := new(repository_gameseeks_mock.GameseeksMockRepo)

	mockRepo.On("Insert", mock.AnythingOfType("*domain.Gameseek")).Return(nil).Once()

	r := httprouter.New()
	r = NewGameseeksHandler(r, mockRepo)

	reqBody, err := json.Marshal(mockGameseek)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/gameseeks", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockRepo.AssertExpectations(t)
}
