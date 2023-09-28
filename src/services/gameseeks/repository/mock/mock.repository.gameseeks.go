package repository_gameseeks_mock

import (
	"context"

	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	services_database "github.com/lookingcoolonavespa/go_crochess_backend/src/services/database"
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

func (c *GameseeksMockRepo) Insert(ctx context.Context, g *domain.Gameseek) error {
	args := c.Called(ctx, g)

	return args.Error(0)
}

func (c *GameseeksMockRepo) Delete(ctx context.Context, db services_database.DBExecutor, seekers ...string) error {
	var interfaceSeekers []interface{}
	for _, s := range seekers {
		interfaceSeekers = append(interfaceSeekers, s)
	}
	args := c.Called(ctx, db, seekers[0], seekers[1])

	return args.Error(0)
}
