package message

type Message struct {
	Id      string `json:"id"`
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
	Order   int64  `json:"order"`
}

type MessagesByTopic map[string][]Message

func (m MessagesByTopic) AddMessage(topic string, message Message) {
	message.Order = m.GetLatestOrder()
	m[topic] = append(m[topic], message)
}

func (m MessagesByTopic) GetLatestOrder() int64 {
	var max int64 = -1
	for _, messages := range m {
		for _, message := range messages {
			if message.Order > max {
				max = message.Order
			}
		}
	}

	return max
}
