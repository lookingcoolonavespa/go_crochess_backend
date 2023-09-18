package domain

type Game struct {
	ID                   int         `json:"id"`
	WhiteID              string      `json:"white_id"`
	BlackID              string      `json:"black_id"`
	Time                 int         `json:"time"`
	Increment            int         `json:"increment"`
	TimeStampAtTurnStart int64       `json:"time_stamp_at_turn_start"`
	WhiteTime            int         `json:"white_time"`
	BlackTime            int         `json:"black_time"`
	History              string      `json:"history"`
	Moves                string      `json:"moves"`
	Result               string      `json:"result"`
	Method               string      `json:"method"`
	Version              int         `json:"version"`
	DrawRecord           *DrawRecord `json:"draw_record"`
}

type GameRepo interface {
	Get(id int) (*Game, error)
	Update(id int, changes map[string]interface{}) error
	Insert(g *Game) error
}
