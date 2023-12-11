package main

import (
	"encoding/json"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type hubData struct {
	Client   *Client
	GameId   uuid.UUID `json:"gameId"`
	Nickname string    `json:"nickname"`
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool
	games   map[*Game]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	create chan *hubData
	join   chan *hubData
	start  chan *hubData
	kick   chan *hubData
	leave  chan *hubData
	play   chan *gameData

	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		create: make(chan *hubData),
		join:   make(chan *hubData),
		start:  make(chan *hubData),
		kick:   make(chan *hubData),
		leave:  make(chan *hubData),
		play:   make(chan *gameData),

		broadcast:  make(chan []byte),
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
	msg, _ := json.Marshal(game)
	for _, player := range game.Players {
		player.Client.send <- msg
	}
}

func (d hubData) sendMessage(msg string) {
	d.Client.send <- []byte(msg)
}

func (h *Hub) run() {
	for {
	selectLoop:
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

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case data := <-h.create:
			game := newGame(data.GameId)

			player := newPlayer(data.Nickname, data.Client)
			player.Rank = Host
			game.Players = append(game.Players, *player)
			h.games[game] = false
			log.WithField("GameInfo", game).Info("Game created")
			data.sendMessage("create ok")
			h.sendGameStatus(game)

		case data := <-h.join:
			for game := range h.games {
				if !h.isGameStarted(game) && game.Id == data.GameId {
					for _, player := range game.Players {
						if player.Nickname == data.Nickname {
							data.sendMessage("join ko")
							break selectLoop
						}
					}

					player := newPlayer(data.Nickname, data.Client)
					game.Players = append(game.Players, *player)
					log.WithField("GameInfo", game).Info("Player joined")

					data.sendMessage("join ok")

					h.sendGameStatus(game)
					break
				}
			}

		case data := <-h.kick:
			for game := range h.games {
				if !h.isGameStarted(game) && game.Id == data.GameId {
					for i, player := range game.Players {
						if player.Nickname == data.Nickname {
							game.Players = append(game.Players[:i], game.Players[i+1:]...)
							log.WithField("GameInfo", game).Info("Player kicked")
							data.sendMessage("kick ok")
							break selectLoop
						}
					}

					h.sendGameStatus(game)
				}
			}

		case data := <-h.leave:
			for game := range h.games {
				if game.Id == data.GameId {
					var host = false

					for i, player := range game.Players {
						if player.Client == data.Client && player.Rank == Host {
							host = true
							break
						}

						// Leave
						if player.Client == data.Client {
							game.Players = append(game.Players[:i], game.Players[i+1:]...)
							log.WithField("GameInfo", game).Info("Player leaved")
							data.sendMessage("game leave ok")
							break
						}
					}

					// Kick All and destroy game
					if host {
						for _, player := range game.Players {
							player.Client.send <- []byte("game closed ok")
						}
						game.Players = nil
						delete(h.games, game)
						log.WithField("GameId", data.GameId).Info("Game destroy")
					}
				}
			}

		case data := <-h.start:
			for game := range h.games {
				if game.Id == data.GameId && !h.isGameStarted(game) {
					for _, player := range game.Players {
						if player.Client == data.Client && player.Rank == Host {
							game.start(data)
							data.sendMessage("start ok")
							log.WithField("GameInfo", game).Info("Game started")

							h.sendGameStatus(game)
							h.games[game] = true
							break
						}
					}
				}
			}

		case data := <-h.play:
			for game := range h.games {
				if game.Id == data.GameId && h.isGameStarted(game) {
					game.play(data)
				}
			}
		}
	}
}
