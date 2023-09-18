package repository_game_mock

import (
	"github.com/lookingcoolonavespa/go_crochess_backend/domain"
	"github.com/stretchr/testify/mock"
)

type GameMockRepo struct {
	mock.Mock
}

func (c *GameMockRepo) Get(id int) (*domain.Game, error) {
	args := c.Called(id)
	result := args.Get(0)

	if result == nil {
		return nil, args.Error(1)
	}

	return result.(*domain.Game), args.Error(1)
}

func (c *GameMockRepo) Update(id int, changes map[string]interface{}) error {
	args := c.Called(id, changes)

	return args.Error(0)
}

func (c *GameMockRepo) Insert(g *domain.Game) error {
	args := c.Called(g)

	return args.Error(0)
}
