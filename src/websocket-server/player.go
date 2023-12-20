package main

type Rank int

const (
	Host Rank = iota
	Guest
)

type Player struct {
	Nickname string  `json:"nickname"`
	Rank     Rank    `json:"rank"`
	Role     Role    `json:"-"`
	Client   *Client `json:"-"`
	Position int     `json:"position"`
}

type Role int

const (
	Undercover Role = iota
	White
	Civilian
)

func newPlayer(nickname string, client *Client) *Player {
	return &Player{
		Nickname: nickname,
		Rank:     Guest,
		Client:   client,
		Role:     Civilian,
	}
}
