package mock_usecase_gameseeks

import (
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

type GameseeksMockUseCase struct {
	mock.Mock
}

func (c *GameseeksMockUseCase) OnAccept(
	ctx context.Context,
	g domain.Game,
	r domain.Room,
) (int, error) {
	args := c.Called(ctx, g, r)
	id := args.Get(0)

	return id.(int), args.Error(1)
}
