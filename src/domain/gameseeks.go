package domain

import (
	"context"
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
	Insert(context.Context, Gameseek) error
	DeleteFromSeeker(context.Context, string) error
}

type GameseeksUseCase interface {
	OnAccept(ctx context.Context, g Game) (int, error)
}
