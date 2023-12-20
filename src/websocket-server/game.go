package main

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Action int

const (
	Nothing Action = iota
	WriteDown
	Vote
)

type Game struct {
	Id        uuid.UUID `json:"gameId"`
	Word      string    `json:"-"`
	Players   []Player  `json:"players"`
	Turn      int       `json:"turn"`
	Votes     []string  `json:"-"`
	Action    Action    `json:"action"`
	Initiator Player    `json:"initiator"`
}

type gameData struct {
	hubData
	Command string
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
		for i, player := range g.Players {
			if player.Client == data.Client && g.Votes[i] != "" {
				g.Votes[i] = data.Command
				info := newInfo(g.Votes[i])
				info.GameInfo = *g
				info.GameInfo.Action = Vote
				info.GameInfo.Initiator = player
				successResult := Response{Info: *info}
				player.Client.sendResponse(successResult)
			}
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
			log.WithField("maxVote", maxVote).Info("Vote")
		}
	}

	// Write Down word
	for i, player := range g.Players {
		if player.Client == data.Client && i == g.Turn {
			info := newInfo(data.Command)
			info.GameInfo = *g
			info.GameInfo.Action = WriteDown
			info.GameInfo.Initiator = player
			g.Turn++
			g.handleTurn(*info)

			if g.Turn == len(g.Players)+1 {
				g.Votes = make([]string, len(g.Players))
			}
			break
		}
	}
}

func (g *Game) handleTurn(info InfoResponse) {
	info.GameInfo.Turn = g.Turn
	successResult := Response{Info: info}
	for _, p := range g.Players {
		p.Client.sendResponse(successResult)
	}
}

func (g *Game) start(data *hubData) {
	for i, _ := range g.Players {
		g.Players[i].Position = i
	}

	g.Word = "a"
	g.Turn = 0

	randomUnderCover, _ := genRandNum(0, len(g.Players))
	g.Players[randomUnderCover].Role = Undercover
	//randomWhite := randomUnderCover
	//for randomWhite == randomUnderCover {
	//	randomWhite, _ = genRandNum(1, len(g.Players))
	//}
	//g.Players[randomWhite].Role = White

	info := newInfo("")
	info.GameInfo = *g
	g.handleTurn(*info)
}
