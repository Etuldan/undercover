package main

import (
	"encoding/json"
)

type ErrorCode int

const (
	GameNotFound ErrorCode = iota
	GameNotAvailable
	NicknameNotAvailable
)

type ErrorResponse struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (client *Client) sendError(err ErrorResponse) {
	msg, _ := json.Marshal(err)
	client.send <- []byte(msg)
}

func newErr(code ErrorCode, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}
