package mock_usecase_game

import (
	"context"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/stretchr/testify/mock"
)

type MockGameUseCase struct {
	mock.Mock
}

func (c *MockGameUseCase) Get(ctx context.Context, id int) (domain.Game, error) {
	args := c.Called(ctx, id)
	res := args.Get(0)

	return res.(domain.Game), args.Error(1)
}

func (c *MockGameUseCase) UpdateOnMove(
	ctx context.Context,
	gameID int,
	playerID string,
	move string,
	_ func(domain.GameChanges),
) (domain.GameChanges, bool, error) {
	args := c.Called(ctx, gameID, playerID, move)
	changes := args.Get(0)
	updated := args.Get(1)

	return changes.(domain.GameChanges), updated.(bool), args.Error(2)
}

func (c *MockGameUseCase) UpdateDraw(
	ctx context.Context,
	gameID int,
	whiteDrawStatus bool,
	blackDrawStatus bool,
) (domain.GameChanges, bool, error) {
	args := c.Called(ctx, gameID, whiteDrawStatus, blackDrawStatus)
	changes := args.Get(0)
	updated := args.Get(1)

	return changes.(domain.GameChanges), updated.(bool), args.Error(2)
}

func (c *MockGameUseCase) UpdateResult(
	ctx context.Context,
	gameID int,
	method string,
	result string,
) (domain.GameChanges, bool, error) {
	args := c.Called(ctx, gameID, method, result)
	changes := args.Get(0)
	updated := args.Get(1)

	return changes.(domain.GameChanges), updated.(bool), args.Error(2)
}
