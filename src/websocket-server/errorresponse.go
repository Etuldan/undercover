package main

type ErrorCode int

const (
	NoError ErrorCode = iota
	GameNotFound
	GameNotAvailable
	NicknameNotAvailable
	IncorrectGameState
	PlayerNotFound
	InsufficientPermission
)

type ErrorResponse struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func newErr(code ErrorCode, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}
