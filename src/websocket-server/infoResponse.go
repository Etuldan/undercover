package main

type InfoResponse struct {
	Message  string `json:"message"`
	GameInfo Game   `json:"gameInfo"`
}

func newInfo(message string) *InfoResponse {
	return &InfoResponse{
		Message: message,
	}
}
