package main

import (
	"encoding/json"
)

type Response struct {
	Error ErrorResponse `json:"error"`
	Info  InfoResponse  `json:"message"`
}

func (client *Client) sendResponse(err Response) {
	msg, _ := json.Marshal(err)
	client.send <- []byte(msg)
}

func newResponse(errorResponse ErrorResponse, infoResponse InfoResponse) *Response {
	return &Response{
		Error: errorResponse,
		Info:  infoResponse,
	}
}
