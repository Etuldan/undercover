package main

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Action int

const (
	Nothing Action = iota
	WriteDown
	Vote
	Eliminated
	DisplayWord
	WhiteGuess
	Winner
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
					result := Response{Info: *info}
					player.Client.sendResponse(result)
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
						g.Action = WhiteGuess
						log.WithField("GameInfo", g).Info("Mr White last chance")
						// TODO
					} else if player.Role == Undercover {
						log.WithField("GameInfo", g).Info("Undercover eliminated")
						info := newInfo("undercover")
						info.GameInfo = *g
						info.GameInfo.Action = Eliminated
						result := Response{Info: *info}
						player.Client.sendResponse(result)
						g.checkEndOfGame()
					} else {
						log.WithField("GameInfo", g).Info("Civilian eliminated")
						info := newInfo("civilian")
						info.GameInfo = *g
						info.GameInfo.Action = Eliminated
						result := Response{Info: *info}
						player.Client.sendResponse(result)
						g.checkEndOfGame()
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
			} else if g.Action == WhiteGuess {
				// TODO
				log.WithField("GameInfo", g).WithField("Word", data.Command).Info("White Guess")
				if g.Word == data.Command {
					log.WithField("GameInfo", g).Info("Game End : White Wins")
					info := newInfo(g.Word)
					info.GameInfo = *g
					info.GameInfo.Action = Winner
					result := Response{Info: *info}
					player.Client.sendResponse(result)
				} else {
					log.WithField("GameInfo", g).Info("White Eliminated")
					info := newInfo("")
					info.GameInfo = *g
					info.GameInfo.Action = Eliminated
					result := Response{Info: *info}
					player.Client.sendResponse(result)

					g.Turn = 0
					g.handleTurn(*info)
				}
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
	result := Response{Info: info}
	for _, p := range g.Players {
		p.Client.sendResponse(result)
	}
}

func (g *Game) checkEndOfGame() {
	// TODO
	countCivilian := 0
	countUndercover := 0
	countWhite := 0
	for _, player := range g.Players {
		if !player.Eliminated {
			if player.Role == White {
				countWhite++
			} else if player.Role == Undercover {
				countUndercover++
			} else if player.Role == Civilian {
				countCivilian++
			}
		}
	}
	if countWhite == 0 && countUndercover == 0 {
		log.WithField("GameInfo", g).Info("Game End : Civilian Wins")
		// Civilian WIN
	}
	if countCivilian == 1 {
		log.WithField("GameInfo", g).Info("Game End : Undercover & MrWhite Wins")
		// Undercover & White WIN
	}
}

func (g *Game) start(data *hubData) {
	g.Turn = 0

	r := rand.New(rand.NewSource(time.Now().Unix()))
	for j, i := range r.Perm(len(g.Players)) {
		g.Players[i].Position = j
	}

	// TODO : Randomize word
	g.Word = "Word"
	synonym := "Synonym"

	// TODO : Configurable number of Undercover & White
	randomUnderCover := r.Intn(len(g.Players)) // Random from 0 to Max
	g.Players[randomUnderCover].Role = Undercover
	randomWhite := randomUnderCover
	for randomWhite == randomUnderCover {
		randomWhite = r.Intn(len(g.Players)-1) + 1 // Random from 1 to Max
	}
	g.Players[randomWhite].Role = White

	for _, player := range g.Players {
		if player.Role == Civilian {
			info := newInfo(g.Word)
			info.GameInfo = *g
			info.GameInfo.Action = DisplayWord
			result := Response{Info: *info}
			player.Client.sendResponse(result)
		} else if player.Role == Undercover {
			info := newInfo(synonym)
			info.GameInfo = *g
			info.GameInfo.Action = DisplayWord
			result := Response{Info: *info}
			player.Client.sendResponse(result)
		} else if player.Role == White {
			info := newInfo("")
			info.GameInfo = *g
			info.GameInfo.Action = DisplayWord
			result := Response{Info: *info}
			player.Client.sendResponse(result)
		}
	}

	log.WithField("GameInfo", g).WithField("Undercover", g.Players[randomUnderCover].Nickname).WithField("Word", g.Word).Info("Game Initiated")
}
