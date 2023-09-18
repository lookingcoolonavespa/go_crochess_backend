package repository_gameseeks_mock

import (
	"github.com/lookingcoolonavespa/go_crochess_backend/domain"
	"github.com/stretchr/testify/mock"
)

type GameseeksMockRepo struct {
	mock.Mock
}

func (c *GameseeksMockRepo) List() ([]domain.Gameseek, error) {
	args := c.Called()
	result := args.Get(0)

	return result.([]domain.Gameseek), args.Error(1)
}

func (c *GameseeksMockRepo) Insert(g *domain.Gameseek) error {
	args := c.Called(g)

	return args.Error(0)
}

func (c *GameseeksMockRepo) Delete(seekers ...string) error {
	args := c.Called(seekers)

	return args.Error(1)
}
