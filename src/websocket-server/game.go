package main

import (
	"fmt"

	"github.com/google/uuid"
)

type Game struct {
	Id      uuid.UUID `json:"gameId"`
	Word    string    `json:"-"`
	Players []Player  `json:"players"`
	Turn    int       `json:"turn"`
	Votes   []string  `json:"-"`
}

type gameData struct {
	hubData
	Word string
	Vote string
}

func newGame(idGame uuid.UUID) *Game {
	return &Game{
		Id:      idGame,
		Players: make([]Player, 0),
	}
}

func (g *Game) play(data *gameData) {
	// Vote Time !
	if g.Turn > len(g.Players) {
		index := -1
		for i, player := range g.Players {
			if player.Client == data.Client {
				index = i
			}
		}
		if index != -1 {
			g.Votes[index] = data.Vote
		}
		everyoneVote := true
		for _, value := range g.Votes {
			if value == "" {
				everyoneVote = false
			}
		}
		if everyoneVote {
			dict := make(map[string]int)
			for _, vote := range g.Votes {
				dict[vote]++
			}
			maxValue := 0
			maxVote := ""
			for vote, value := range dict {
				if value > maxValue {
					maxValue = value
					maxVote = vote
				}
			}
			fmt.Println("max vote", maxVote)
		}
	}

	// Write Down word
	for i, player := range g.Players {
		if player.Client == data.Client && i == g.Turn {
			for _, p := range g.Players {
				p.Client.send <- []byte(data.Word)
			}
			g.Turn++
			if g.Turn == len(g.Players)+1 {
				g.Votes = make([]string, len(g.Players))
			}
			break
		}
	}
}

func (g *Game) start(data *hubData) {
	g.Word = "a"

	g.Turn = 0

	random1, _ := genRandNum(0, len(g.Players))
	random2 := random1
	for random2 == random1 {
		random2, _ = genRandNum(1, len(g.Players))
	}
	g.Players[random1].Role = Undercover
	g.Players[random2].Role = White

	fmt.Println("Game Start")
}
