package main

import (
	"encoding/json"
)

type Response struct {
	Error    ErrorResponse `json:"error"`
	Info     InfoResponse  `json:"info"`
	GameInfo Game          `json:"gameInfo"`
}

func (client *Client) sendResponse(response Response) {
	msg, _ := json.Marshal(response)
	client.send <- []byte(msg)
}
