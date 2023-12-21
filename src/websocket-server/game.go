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
	Eliminated
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
	if g.Turn == len(g.Players) {
		for i, player := range g.Players {
			if player.Client == data.Client {
				if g.Votes[i] != "" {
					err := newErr(NotYourTurn, "You already voted")
					result := Response{Error: *err}
					data.Client.sendResponse(result)
					return
				} else {
					log.WithField("GameInfo", g).WithField("Vote", data.Command).Info("New Vote")
					g.Votes[i] = data.Command
					info := newInfo(g.Votes[i])
					info.GameInfo = *g
					info.GameInfo.Action = Vote
					info.GameInfo.Initiator = player
					successResult := Response{Info: *info}
					player.Client.sendResponse(successResult)
				}
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
				} else if value == maxValue {
					// TODO Random
				}
			}
			log.WithField("GameInfo", g).WithField("Vote", maxVote).WithField("NbVote", maxValue).Info("Vote Result")

			for i, player := range g.Players {
				if player.Nickname == maxVote {
					g.Players[i].Eliminated = true
					if player.Role == White {
						g.Turn = player.Position
						log.WithField("GameInfo", g).Info("Mr White last chance")
						// TODO
					} else if player.Role == Undercover {
						log.WithField("GameInfo", g).Info("Undercover eliminated")
						// TODO
					} else {
						log.WithField("GameInfo", g).Info("Civilian eliminated")
						// TODO
					}
				}
			}

			info := newInfo(maxVote)
			info.GameInfo = *g
			info.GameInfo.Action = Eliminated
			g.Turn = 0
			g.handleTurn(*info)
			return
		}

		return
	}

	// Write Down word
	for i, player := range g.Players {
		if player.Client == data.Client {
			if i != g.Turn {
				err := newErr(NotYourTurn, "Not your turn")
				result := Response{Error: *err}
				data.Client.sendResponse(result)
				return
			} else {
				log.WithField("GameInfo", g).WithField("Word", data.Command).Info("Word")
				info := newInfo(data.Command)
				info.GameInfo = *g
				info.GameInfo.Action = WriteDown
				info.GameInfo.Initiator = player
				g.Turn++
				g.handleTurn(*info)

				if g.Turn == len(g.Players) {
					g.Votes = make([]string, len(g.Players))
				}
				return
			}
		}
	}
	err := newErr(PlayerNotFound, "Invalid player")
	result := Response{Error: *err}
	data.Client.sendResponse(result)
}

func (g *Game) handleTurn(info InfoResponse) {
	//info.GameInfo.Turn = g.Turn
	successResult := Response{Info: info}
	for _, p := range g.Players {
		p.Client.sendResponse(successResult)
	}
}

func (g *Game) start(data *hubData) {
	g.Turn = 0

	// TODO : Randomize order
	for i, _ := range g.Players {
		g.Players[i].Position = i
	}

	// TODO : Randomize word
	g.Word = "a"

	// TODO : Configurable number of Undercover & White
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
	log.WithField("GameInfo", g).WithField("Undercover", g.Players[randomUnderCover].Nickname).WithField("Word", g.Word).Info("New Game")
}
