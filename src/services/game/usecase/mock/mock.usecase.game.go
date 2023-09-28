package mock_usecase_game

import (
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

type GameMockUseCase struct {
	mock.Mock
}

func (c *GameMockUseCase) Get(ctx context.Context, gameID int) (*domain.Game, error) {
	args := c.Called(gameID)
	result := args.Get(0)

	if result == nil {
		return nil, args.Error(1)
	}

	return result.(*domain.Game), args.Error(1)
}

func (c *GameMockUseCase) Start(ctx context.Context, g *domain.Game) (int64, error) {
	args := c.Called(ctx, g)
	res := args.Get(0)

	return res.(int64), args.Error(1)
}

func (c *GameMockUseCase) UpdateOnMove(
	ctx context.Context,
	gameID int,
	playerID string,
	move string,
) error {
	args := c.Called(ctx, gameID, playerID, move)

	return args.Error(0)
}
