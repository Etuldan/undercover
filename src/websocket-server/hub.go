package main

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type hubData struct {
	Client   *Client
	GameId   uuid.UUID `json:"gameId"`
	Nickname string    `json:"nickname"`
}

type Hub struct {
	cfg Config
	// Registered clients.
	clients map[*Client]bool
	games   map[*Game]bool

	create chan *hubData
	join   chan *hubData
	start  chan *hubData
	kick   chan *hubData
	leave  chan *hubData
	play   chan *gameData

	register   chan *Client
	unregister chan *Client
}

func newHub(config Config) *Hub {
	return &Hub{
		cfg:    config,
		create: make(chan *hubData),
		join:   make(chan *hubData),
		start:  make(chan *hubData),
		kick:   make(chan *hubData),
		leave:  make(chan *hubData),
		play:   make(chan *gameData),

		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		games:      make(map[*Game]bool),
	}
}

func (h *Hub) isGameStarted(game *Game) bool {
	return h.games[game]
}

func (h *Hub) sendGameStatus(game *Game) {
	info := newInfo("status")
	info.GameInfo = *game
	successResult := Response{Info: *info}
	for _, player := range game.Players {
		player.Client.sendResponse(successResult)
	}
}

func (h *Hub) closeGame(game *Game) {
	log.WithField("GameInfo", game).Info("Close Game")

	info := newInfo("closed")
	info.GameInfo = *game
	info.GameInfo.Action = Closed
	successResult := Response{Info: *info}
	for _, player := range game.Players {
		player.Client.sendResponse(successResult)
	}
	game.Players = nil
	for currGame := range h.games {
		if currGame.Id == game.Id {
			delete(h.games, currGame)
		}
	}
}

func caseLoop(h *Hub) {
	select {
	case client := <-h.register:
		h.clients[client] = true
	case client := <-h.unregister:
		if _, ok := h.clients[client]; ok {
			for game := range h.games {
				for i, player := range game.Players {
					if player.Client == client {
						game.Players = append(game.Players[:i], game.Players[i+1:]...)
					}
				}
			}
			delete(h.clients, client)
			close(client.send)
		}

	case data := <-h.create:
		if h.cfg.Debug.GameUuid != "" {
			data.GameId = uuid.MustParse(h.cfg.Debug.GameUuid)
		} else {
			data.GameId = uuid.New()
		}
		game := newGame(data.GameId, h)

		player := newPlayer(data.Nickname, data.Client)
		player.Rank = Host
		game.Players = append(game.Players, *player)
		h.games[game] = false
		log.WithField("GameInfo", game).Info("Game created")
		info := newInfo("Game created")
		info.GameInfo = *game
		result := Response{Info: *info}
		data.Client.sendResponse(result)

	case data := <-h.join:
		for game := range h.games {
			if game.Id == data.GameId {
				if h.isGameStarted(game) {
					err := newErr(IncorrectGameState, "Game already started")
					result := Response{Error: *err}
					data.Client.sendResponse(result)
					return
				} else {
					for _, player := range game.Players {
						if player.Nickname == data.Nickname {
							err := newErr(NicknameNotAvailable, "Nickname already taken")
							result := Response{Error: *err}
							data.Client.sendResponse(result)
							return
						}
					}

					player := newPlayer(data.Nickname, data.Client)
					game.Players = append(game.Players, *player)
					log.WithField("GameInfo", game).Info("Player joined")

					info := newInfo("Game joined")
					info.GameInfo = *game
					result := Response{Info: *info}
					data.Client.sendResponse(result)

					h.sendGameStatus(game)
					return
				}
			}
		}
		err := newErr(GameNotFound, "Game not found")
		result := Response{Error: *err}
		data.Client.sendResponse(result)
		return

	case data := <-h.kick:
		for game := range h.games {
			if game.Id == data.GameId {
				if h.isGameStarted(game) {
					err := newErr(IncorrectGameState, "Game already started")
					result := Response{Error: *err}
					data.Client.sendResponse(result)
					return
				} else {
					for i, player := range game.Players {
						if player.Nickname == data.Nickname {
							game.Players = append(game.Players[:i], game.Players[i+1:]...)
							log.WithField("GameInfo", game).Info("Player kicked")

							info := newInfo("Player kicked")
							info.GameInfo = *game
							result := Response{Info: *info}
							data.Client.sendResponse(result)
							h.sendGameStatus(game)
							return
						}
					}
					err := newErr(PlayerNotFound, "Player not found")
					result := Response{Error: *err}
					data.Client.sendResponse(result)
					return
				}
			}
		}
		err := newErr(GameNotFound, "Game not found")
		result := Response{Error: *err}
		data.Client.sendResponse(result)
		return

	case data := <-h.leave:
		for game := range h.games {
			if game.Id == data.GameId {
				host := false

				for i, player := range game.Players {
					// Leave
					if player.Client == data.Client {
						if player.Rank == Host {
							host = true
							break
						} else {
							game.Players = append(game.Players[:i], game.Players[i+1:]...)
							log.WithField("GameInfo", game).Info("Player leaved")
							info := newInfo("Game Leaved")
							info.GameInfo = *game
							result := Response{Info: *info}
							data.Client.sendResponse(result)
							return
						}
					}
				}

				// Kick All and destroy game
				if host {
					for _, player := range game.Players {
						info := newInfo("Game destroyed")
						info.GameInfo = *game
						result := Response{Info: *info}
						player.Client.sendResponse(result)
					}
					game.Players = nil
					delete(h.games, game)
					log.WithField("GameId", data.GameId).Info("Game destroyed")

					info := newInfo("Game destroyed")
					info.GameInfo = *game
					result := Response{Info: *info}
					data.Client.sendResponse(result)
					return
				}
			}
		}
		err := newErr(GameNotFound, "Game not found")
		result := Response{Error: *err}
		data.Client.sendResponse(result)
		return

	case data := <-h.start:
		for game := range h.games {
			if game.Id == data.GameId {
				if h.isGameStarted(game) {
					err := newErr(IncorrectGameState, "Game already started")
					result := Response{Error: *err}
					data.Client.sendResponse(result)
					return
				} else {
					for _, player := range game.Players {
						if player.Client == data.Client {
							if player.Rank != Host {
								err := newErr(InsufficientPermission, "You are not permitted to start the game")
								result := Response{Error: *err}
								data.Client.sendResponse(result)
								return
							} else {
								info := newInfo("Game started")
								info.GameInfo = *game
								result := Response{Info: *info}
								data.Client.sendResponse(result)

								log.WithField("GameInfo", game).Info("Game started")
								h.sendGameStatus(game)
								h.games[game] = true

								game.start(data)

								return
							}
						}
					}
				}
			}
		}
		err := newErr(GameNotFound, "Game not found")
		result := Response{Error: *err}
		data.Client.sendResponse(result)
		return

	case data := <-h.play:
		for game := range h.games {
			if game.Id == data.GameId {
				if !h.isGameStarted(game) {
					err := newErr(IncorrectGameState, "Game not started")
					result := Response{Error: *err}
					data.Client.sendResponse(result)
					return
				} else {
					game.play(data)
					return
				}
			}
		}

		err := newErr(GameNotFound, "Game not found")
		result := Response{Error: *err}
		data.Client.sendResponse(result)
		return
	}
}

func (h *Hub) run() {
	for {
		caseLoop(h)
	}
}
