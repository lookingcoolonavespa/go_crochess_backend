package repository_game_mock

import (
	"context"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/utils"
	"github.com/stretchr/testify/mock"
)

type GameMockRepo struct {
	mock.Mock
}

func (c *GameMockRepo) Get(ctx context.Context, id int) (domain.Game, error) {
	args := c.Called(ctx, id)
	result := args.Get(0)

	return result.(domain.Game), args.Error(1)
}

func (c *GameMockRepo) Update(
	ctx context.Context,
	id int,
	version int,
	changes utils.Changes[domain.GameFieldJsonTag],
) (bool, error) {
	args := c.Called(ctx, id, version, changes)
	result := args.Get(0)

	return result.(bool), args.Error(1)
}

func (c *GameMockRepo) Insert(
	ctx context.Context,
	g domain.Game,
) (int, error) {
	args := c.Called(ctx, g)
	gameID := args.Get(0)

	return gameID.(int), args.Error(1)
}
