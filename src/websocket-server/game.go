package main

import (
	"encoding/csv"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Action int

const (
	Nothing Action = iota
	WriteDown
	Vote
	Voted
	Eliminated
	DisplayWord
	WhiteGuess
	Winner
	Closed
)

type Game struct {
	Hub     *Hub       `json:"-"`
	Id      uuid.UUID  `json:"gameId"`
	Word    string     `json:"-"`
	Players []Player   `json:"players"`
	Turn    int        `json:"turn"`
	Votes   []string   `json:"-"`
	logger  *log.Entry `json:"-"`
	Action  Action     `json:"action"`
}

type gameData struct {
	hubData
	Command string
}

func newGame(idGame uuid.UUID, hub *Hub) *Game {
	return &Game{
		Hub:     hub,
		Id:      idGame,
		Players: make([]Player, 0),
		logger:  log.WithField("gameId", idGame),
	}
}

func (g *Game) play(data *gameData) {
	logger := g.logger.WithField("GameInfo", g)
	// Vote Time !
	if g.Turn == len(g.Players) {
		for i, player := range g.Players {
			if player.Client == data.Client {
				if g.Votes[i] != "" {
					err := newErr(NotYourTurn, "You already voted")
					result := Response{Error: *err, GameInfo: *g}
					data.Client.sendResponse(result)
					return
				} else {
					logger.WithField("Vote", data.Command).Info("New Vote")
					g.Votes[i] = data.Command
					info := newInfo(g.Votes[i])
					info.Action = Vote
					info.Initiator = player
					result := Response{Info: *info, GameInfo: *g}
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
			maxVoteSlice := make([]string, 0)
			for vote, value := range dict {
				if value > maxValue {
					maxValue = value
					maxVoteSlice = []string{vote}
				} else if value == maxValue {
					maxVoteSlice = append(maxVoteSlice, vote)
				}
			}
			r := rand.New(rand.NewSource(time.Now().Unix()))
			maxVote := maxVoteSlice[r.Intn(len(maxVoteSlice))]
			logger.WithField("Vote", maxVote).WithField("NbVote", maxValue).Info("Vote Result")

			info := newInfo(maxVote)
			info.Action = Voted
			result := Response{Info: *info, GameInfo: *g}
			for _, player := range g.Players {
				player.Client.sendResponse(result)
			}

			for i, player := range g.Players {
				if player.Nickname == maxVote {
					g.Players[i].Eliminated = true
					if player.Role == White {
						g.Turn = player.Position
						g.Action = WhiteGuess
						info.Action = WhiteGuess
						logger.Info("Mr White last chance")
						info := newInfo("")
						g.handleTurn(*info)
						return
					} else if player.Role == Undercover {
						logger.Info("Undercover eliminated")
						g.Turn = 0
						info := newInfo("undercover")
						info.Action = Eliminated
						g.handleTurn(*info)
						g.checkEndOfGame()
						return
					} else {
						logger.Info("Civilian eliminated")
						g.Turn = 0
						info := newInfo("civilian")
						info.Action = Eliminated
						g.handleTurn(*info)
						g.checkEndOfGame()
						return
					}
				}
			}
		}

		return
	}

	// Write Down word
	for _, player := range g.Players {
		if player.Client == data.Client {
			if player.Position != g.Turn {
				err := newErr(NotYourTurn, "Not your turn")
				result := Response{Error: *err, GameInfo: *g}
				data.Client.sendResponse(result)
				return
			} else if g.Action == WhiteGuess {
				logger.WithField("Word", data.Command).Info("White Guess")
				if g.Word == data.Command {
					logger.Info("Game End : White Wins")
					info := newInfo(g.Word)
					info.Action = Winner
					result := Response{Info: *info, GameInfo: *g}
					player.Client.sendResponse(result)
					g.Hub.closeGame(g)
					return
				} else {
					logger.Info("White Eliminated")
					info := newInfo("")
					info.Action = Eliminated
					result := Response{Info: *info, GameInfo: *g}
					player.Client.sendResponse(result)
					g.Turn = 0
					g.handleTurn(*info)
					return
				}
			} else {
				logger.WithField("Word", data.Command).Info("Word")
				info := newInfo(data.Command)
				info.Action = WriteDown
				info.Initiator = player
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
	result := Response{Error: *err, GameInfo: *g}
	data.Client.sendResponse(result)
}

func (g *Game) handleTurn(info InfoResponse) {
	result := Response{Info: info, GameInfo: *g}
	for _, p := range g.Players {
		p.Client.sendResponse(result)
	}
}

func (g *Game) checkEndOfGame() {
	// TODO checkEndOfGame
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
		// Civilian WIN
		g.logger.WithField("GameInfo", g).Info("Game End : Civilian Wins")

		g.Hub.closeGame(g)
	}
	if countCivilian == 1 {
		// Undercover & White WIN
		g.logger.WithField("GameInfo", g).Info("Game End : Undercover & MrWhite Wins")

		g.Hub.closeGame(g)
	}
}

func (g *Game) start(data *hubData) {
	logger := g.logger.WithField("GameInfo", g)

	r := rand.New(rand.NewSource(time.Now().Unix()))

	path := "./data/list-words.csv"
	file, err := os.Open(path)
	if err != nil {
		logger.WithError(err).Error("Error while opening the file")
		g.Hub.closeGame(g)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		logger.WithError(err).WithField("file", file).Error("Error while reading the file")
		g.Hub.closeGame(g)
		return
	}
	wordList := map[int][]string{}
	for i, eachrecord := range records {
		wordList[i] = []string{eachrecord[0], eachrecord[1]}
	}
	randomWord := r.Intn(len(wordList))
	randomWordOrder := r.Intn(2) // Random swap left-right
	g.Word = wordList[randomWord][randomWordOrder]
	synonym := wordList[randomWord][(randomWordOrder+1)%2]
	clear(wordList)
	logger.WithField("Word", g.Word).WithField("Synonym", synonym).Info("Word draw")

	g.Turn = 0
	// Randomize positions
	for j, i := range r.Perm(len(g.Players)) {
		g.Players[i].Position = j
	}

	// TODO max player vs number configurable
	listExclude := []int{}
	nbUndercover := g.Hub.cfg.Game.NbUndercover
	for i := 0; i < nbUndercover; i++ {
		random := r.Intn(len(g.Players))
		for checkInList(listExclude, random) {
			random = r.Intn(len(g.Players))
		}
		listExclude = append(listExclude, random)
		g.Players[random].Role = Undercover
		logger.WithField("Undercover", g.Players[random].Nickname).Info("Undercover draw")
	}
	nbWhite := g.Hub.cfg.Game.NbWhite
	for i := 0; i < nbWhite; i++ {
		random := r.Intn(len(g.Players))
		for checkInList(listExclude, random) || g.Players[random].Position == 0 {
			random = r.Intn(len(g.Players))
		}
		listExclude = append(listExclude, random)
		g.Players[random].Role = White
		logger.WithField("White", g.Players[random].Nickname).Info("MrWhite draw")
	}
	clear(listExclude)

	for _, player := range g.Players {
		if player.Role == Civilian {
			info := newInfo(g.Word)
			info.Action = DisplayWord
			result := Response{Info: *info, GameInfo: *g}
			player.Client.sendResponse(result)
		} else if player.Role == Undercover {
			info := newInfo(synonym)
			info.Action = DisplayWord
			result := Response{Info: *info, GameInfo: *g}
			player.Client.sendResponse(result)
		} else if player.Role == White {
			info := newInfo("")
			info.Action = DisplayWord
			result := Response{Info: *info, GameInfo: *g}
			player.Client.sendResponse(result)
		}
	}
}
