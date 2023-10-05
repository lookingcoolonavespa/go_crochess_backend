package repository_gameseeks_mock

import (
	"context"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/stretchr/testify/mock"
)

type GameseeksMockRepo struct {
	mock.Mock
}

func (c *GameseeksMockRepo) List(ctx context.Context) ([]domain.Gameseek, error) {
	args := c.Called(ctx)
	result := args.Get(0)

	return result.([]domain.Gameseek), args.Error(1)
}

func (c *GameseeksMockRepo) Insert(ctx context.Context, g domain.Gameseek) error {
	args := c.Called(ctx, g)

	return args.Error(0)
}

func (c *GameseeksMockRepo) DeleteFromSeeker(ctx context.Context, seeker string) ([]int, error) {
	args := c.Called(ctx, seeker)
	res := args.Get(0)

	return res.([]int), args.Error(1)
}
