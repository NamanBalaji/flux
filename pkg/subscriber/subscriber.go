package subscriber

type Subscriber struct {
	Addr    string
	StartAt int64
}

type SubscribersByTopic map[string][]Subscriber

func NewSubscriber(addr string, latestOrder int64) *Subscriber {
	return &Subscriber{
		Addr:    addr,
		StartAt: latestOrder,
	}
}

func (m SubscribersByTopic) AddSubscriber(topic string, subscriber Subscriber) {
	m[topic] = append(m[topic], subscriber)
}

func (s *Subscriber) PushMessage(topic string) {
	// get latest unAcked message for topic
	// send to subscriber
	// update startAt index
}
