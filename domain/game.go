package domain

type Game struct {
	ID                   int         `json:"id"`
	WhiteID              string      `json:"white_id"`
	BlackID              string      `json:"black_id"`
	Time                 int         `json:"time"`
	Increment            int         `json:"increment"`
	TimeStampAtTurnStart int64       `json:"time_stamp_at_turn_start"`
	WhiteTime            int64       `json:"white_time"`
	BlackTime            int64       `json:"black_time"`
	History              string      `json:"history"`
	Moves                string      `json:"moves"`
	Result               string      `json:"result"`
	Method               string      `json:"method"`
	Version              int         `json:"version"`
	DrawRecord           *DrawRecord `json:"draw_record"`
}

type GameRepo interface {
	Get(g *Game) (*Game, error)
	Update(id int, updater func(g *Game, changes map[string]interface{}) error) error
	Insert(g *Game) error
}
