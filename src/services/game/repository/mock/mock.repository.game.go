package repository_game_mock

import (
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
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

func (c *GameMockRepo) Update(id int, version int, changes map[string]interface{}) (bool, error) {
	args := c.Called(id, version, changes)
	result := args.Get(0)

	return result.(bool), args.Error(1)
}

func (c *GameMockRepo) Insert(g *domain.Game) error {
	args := c.Called(g)

	return args.Error(0)
}
