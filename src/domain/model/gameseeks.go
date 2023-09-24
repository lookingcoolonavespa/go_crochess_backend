package domain

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
	List() ([]Gameseek, error)
	Insert(gs *Gameseek) error
	Delete(seekers ...string) error
}
