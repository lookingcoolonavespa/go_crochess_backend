package mock_usecase_game

import (
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
	"github.com/stretchr/testify/mock"
)

type GameMockUseCase struct {
	mock.Mock
}

func (c *GameMockUseCase) Insert(g *domain.Game) error {
	args := c.Called(g)

	return args.Error(0)
}

func (c *GameMockUseCase) UpdateOnMove(gameID int, playerID string, move string) error {
	args := c.Called(gameID, playerID, move)

	return args.Error(0)
}
