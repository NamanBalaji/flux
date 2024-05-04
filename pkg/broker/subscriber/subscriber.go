package subscriber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/queue"
	"net/http"
	"time"
)

type Subscriber struct {
	Addr         string
	MessageQueue *queue.Queue
	IsActive     bool
}

type MessageResponse struct {
	Message message.Message `json:"message"`
	Topic   string          `json:"topic"`
}

func NewSubscriber(addr string) *Subscriber {
	msgQueue := queue.NewQueue()

	return &Subscriber{
		Addr:         addr,
		MessageQueue: msgQueue,
		IsActive:     true,
	}
}

func (s *Subscriber) WithQueue(q *queue.Queue) *Subscriber {
	s.MessageQueue = q
	return s
}

func (s *Subscriber) AddMessage(msg message.Message) {
	s.MessageQueue.Enqueue(msg)
}

func (s *Subscriber) HandleQueue(topicName string) {
	for {
		if !s.IsActive {
			return
		}

		msg := s.MessageQueue.Peek()
		err := s.pushMessage(msg, topicName)
		if err == nil {
			s.MessageQueue.Dequeue()
		}

	}
}

func (s *Subscriber) pushMessage(msg message.Message, topicName string) error {
	res := MessageResponse{
		Message: msg,
		Topic:   topicName,
	}

	jsonBody, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	// Define retry parameters
	retryCount := 3
	retryInterval := 2 * time.Second

	for i := 0; i < retryCount; i++ {
		resp, err := http.Post(s.Addr, "application/json", bytes.NewBuffer(jsonBody))
		if err == nil && resp.StatusCode == http.StatusOK {
			// Successfully sent the message
			return nil
		}

		// Create a ticker for managing retries
		ticker := time.NewTicker(retryInterval)
		select {
		case <-ticker.C:
			ticker.Stop()
		}
	}

	s.IsActive = false
	return fmt.Errorf("failed to send message after %d retries", retryCount)
}
