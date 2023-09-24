package domain

type DrawRecord struct {
	GameID int  `json:"game_id"`
	White  bool `json:"white"`
	Black  bool `json:"black"`
}
