package mock_usecase_game

import (
	"context"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/lookingcoolonavespa/go_crochess_backend/src/utils"
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
) (utils.Changes, error) {
	args := c.Called(ctx, gameID, playerID, move)
	res := args.Get(0)

	return res.(utils.Changes), args.Error(1)
}
