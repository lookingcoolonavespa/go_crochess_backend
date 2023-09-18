package repository_game_mock

import (
	"github.com/lookingcoolonavespa/go_crochess_backend/domain"
	"github.com/stretchr/testify/mock"
)

type GameMockRepo struct {
	mock.Mock
}

func (c *GameMockRepo) Get(g *domain.Game) (domain.Game, error) {
	args := c.Called(g)
	result := args.Get(0)

	return result.(domain.Game), args.Error(1)
}

func (c *GameMockRepo) Update(g *domain.Game, updater func(g *domain.Game, changes *map[string]interface{})) error {
	args := c.Called(g, updater)

	return args.Error(1)
}

func (c *GameMockRepo) Insert(g *domain.Game) error {
	args := c.Called()

	return args.Error(1)
}
