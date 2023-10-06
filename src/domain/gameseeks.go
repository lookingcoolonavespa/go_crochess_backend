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
	GameseeksRepo interface {
		List(context.Context) ([]Gameseek, error)
		Insert(context.Context, Gameseek) error
		DeleteFromSeeker(context.Context, string) ([]int, error)
	}

	GameseeksUseCase interface {
		OnAccept(
			ctx context.Context,
			g Game,
			onTimeOut func(GameChanges),
		) (gameID int, err error)
	}
)

func (g Gameseek) IsFilled() (bool, []string) {
	missingFields := make([]string, 0)

	if g.Seeker == "" {
		missingFields = append(missingFields, "seeker")
	}
	if g.Time == 0 {
		missingFields = append(missingFields, "time")
	}
	if g.Color == "" {
		missingFields = append(missingFields, "color")
	}

	return len(missingFields) == 0, missingFields
}
