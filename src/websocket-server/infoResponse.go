package main

import (
	"encoding/json"
)

type InfoResponse struct {
	Message string `json:"message"`
}

func (client *Client) sendInfo(err InfoResponse) {
	msg, _ := json.Marshal(err)
	client.send <- []byte(msg)
}

func newInfo(code ErrorCode, message string) *InfoResponse {
	return &InfoResponse{
		Message: message,
	}
}
