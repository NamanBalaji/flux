package handler

type PublishMessageRequest struct {
	Key     string `json:"key"`
	Message string `json:"message"`
	Topic   string `json:"topic"`
}
