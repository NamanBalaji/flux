package handler

type PublishMessageRequest struct {
	Key     string `json:"key"`
	Message string `json:"message"`
	Topic   string `json:"topic"`
}

type ReadMessageRequest struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    int    `json:"offset"`
}
