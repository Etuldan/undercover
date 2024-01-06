package main

type Rank int

const (
	NoRank Rank = iota
	Host
	Guest
)

type Player struct {
	Nickname   string  `json:"nickname"`
	Rank       Rank    `json:"rank"`
	Role       Role    `json:"-"`
	Client     *Client `json:"-"`
	Position   int     `json:"position"`
	Eliminated bool    `json:"eliminated"`
}

type Role int

const (
	NoRole Role = iota
	Undercover
	White
	Civilian
)

func newPlayer(nickname string, client *Client) *Player {
	return &Player{
		Nickname:   nickname,
		Rank:       Guest,
		Client:     client,
		Role:       Civilian,
		Eliminated: false,
	}
}
