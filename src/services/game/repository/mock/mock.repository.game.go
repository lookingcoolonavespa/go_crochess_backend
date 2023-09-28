package repository_game_mock

import (
	"context"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	services_database "github.com/lookingcoolonavespa/go_crochess_backend/src/services/database"
	"github.com/stretchr/testify/mock"
)

type GameMockRepo struct {
	mock.Mock
}

func (c *GameMockRepo) Get(ctx context.Context, db services_database.DBExecutor, id int) (*domain.Game, error) {
	args := c.Called(db, ctx, id)
	result := args.Get(0)

	if result == nil {
		return nil, args.Error(1)
	}

	return result.(*domain.Game), args.Error(1)
}

func (c *GameMockRepo) Update(
	ctx context.Context,
	db services_database.DBExecutor,
	id int,
	version int,
	changes map[string]interface{},
) (bool, error) {
	args := c.Called(db, ctx, id, version, changes)
	result := args.Get(0)

	return result.(bool), args.Error(1)
}

func (c *GameMockRepo) Insert(
	ctx context.Context,
	db services_database.DBExecutor,
	g *domain.Game,
) (int64, error) {
	args := c.Called(ctx, db, g)
	res := args.Get(0)

	return res.(int64), args.Error(1)
}
