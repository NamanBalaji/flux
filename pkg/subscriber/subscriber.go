package subscriber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NamanBalaji/flux/pkg/message"
	"net/http"
)

type Subscriber struct {
	Addr        string
	StartAt     int64
	MessageChan chan message.Message
}

type SubscribersByTopic map[string][]*Subscriber

func NewSubscriber(addr string, latestOrder int64) *Subscriber {
	return &Subscriber{
		Addr:    addr,
		StartAt: latestOrder,
	}
}

func (m SubscribersByTopic) AddSubscriber(topic string, subscriber *Subscriber) {
	if subscriber == nil {
		return
	}

	m[topic] = append(m[topic], subscriber)
}

func (s *Subscriber) PushMessage(msg message.Message) error {
	jsonBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	resp, err := http.Post(s.Addr, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		// handle retries
	}

	return nil
}
