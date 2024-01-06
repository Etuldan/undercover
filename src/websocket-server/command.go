package main

import "github.com/google/uuid"

type Command struct {
	Nickname    string      `json:"nickname"`
	GameId      uuid.UUID   `json:"gameId"`
	CommandCode CommandCode `json:"commandCode"`
	Data        string      `json:"data"`
}

type CommandCode string

const (
	Host   CommandCode = "host"
	Start  CommandCode = "start"
	Join   CommandCode = "join"
	Kick   CommandCode = "kick"
	Play   CommandCode = "play"
	Leave  CommandCode = "leave"
	Status CommandCode = "status"
)
