package handler

type PublishMessageRequest struct {
	Id      string `json:"id"`
	Message string `json:"message"`
	Topic   string `json:"topic"`
}

type RegisterSubscriberRequest struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	ReadOld bool     `json:"readOld"`
}

type UnsubscribeRequest struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
}
