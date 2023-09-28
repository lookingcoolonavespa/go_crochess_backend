package domain

import (
	"context"

	services_database "github.com/lookingcoolonavespa/go_crochess_backend/src/services/database"
)

type (
	Gameseek struct {
		ID        int    `json:"id"`
		Color     string `json:"color"`
		Time      int    `json:"time"`
		Increment int    `json:"increment"`
		Seeker    string `json:"seeker"`
	}
)

type GameseeksRepo interface {
	List(context.Context) ([]Gameseek, error)
	Insert(context.Context, *Gameseek) error
	Delete(ctx context.Context, db services_database.DBExecutor, seekers ...string) error
}

type GameseeksUseCase interface {
	OnAccept(ctx context.Context, g *Game) (int64, error)
}
