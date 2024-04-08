package message

type Message struct {
	Id      string `json:"id"`
	Payload string `json:"payload"`
	Order   int64  `json:"order"`
}

type Messages map[string][]Message

func (m Messages) AddMessage(message Message) {
	_, ok := m[message.Id]
	if !ok {
		m[message.Id] = []Message{message}

		return
	}

	m[message.Id] = append(m[message.Id], message)
}

func (m Messages) GetLatestOrder() int64 {
	var max int64 = 0
	for _, messages := range m {
		for _, message := range messages {
			if message.Order > max {
				max = message.Order
			}
		}
	}

	return max
}
