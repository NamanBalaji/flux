package handler

type PublishMessageRequest struct {
	Id      string `json:"id"`
	Message string `json:"message"`
	Topic   string `json:"topic"`
}
