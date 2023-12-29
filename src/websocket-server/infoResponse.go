package main

type InfoResponse struct {
	Message   string `json:"message"`
	Action    Action `json:"action"`
	Initiator Player `json:"initiator"`
}

func newInfo(message string) *InfoResponse {
	return &InfoResponse{
		Message: message,
	}
}
