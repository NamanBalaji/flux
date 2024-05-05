package subscriber

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/queue"
	"net/http"
	"time"
)

type Subscriber struct {
	Addr         string
	MessageQueue *queue.Queue
	IsActive     bool
	CancelFunc   context.CancelFunc
}

type MessageResponse struct {
	Id      string `json:"id"`
	Payload string `json:"payload"`
	Topic   string `json:"topic"`
}

func NewSubscriber(ctx context.Context, addr string) *Subscriber {
	msgQueue := queue.NewQueue()
	_, cancel := context.WithCancel(ctx)

	return &Subscriber{
		Addr:         addr,
		MessageQueue: msgQueue,
		IsActive:     true,
		CancelFunc:   cancel,
	}
}

func (s *Subscriber) WithQueue(q *queue.Queue) *Subscriber {
	s.MessageQueue = q
	return s
}

func (s *Subscriber) AddMessage(msg *message.Message) {
	s.MessageQueue.Enqueue(msg)
}

func (s *Subscriber) HandleQueue(ctx context.Context, cfg config.Config, topicName string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !s.IsActive {
				return
			}

			msg := s.MessageQueue.Peek()
			err := s.pushMessage(cfg, msg, topicName)
			if err == nil {
				msg = s.MessageQueue.Dequeue()
				msg.Ack(s.Addr)
			}
		}
	}
}

func (s *Subscriber) pushMessage(cfg config.Config, msg *message.Message, topicName string) error {
	res := MessageResponse{
		Id:      msg.Id,
		Payload: msg.Payload,
		Topic:   topicName,
	}

	jsonBody, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	for i := 0; i < cfg.Subscriber.RetryCount; i++ {
		resp, err := http.Post(s.Addr, "application/json", bytes.NewBuffer(jsonBody))
		if err == nil && resp.StatusCode == http.StatusOK {
			// Successfully sent the message
			return nil
		}

		ticker := time.NewTicker(time.Duration(cfg.Subscriber.RetryInterval) * time.Second)
		select {
		case <-ticker.C:
			ticker.Stop()
		}
	}

	s.IsActive = false
	return fmt.Errorf("failed to send message after %d retries", retryCount)
}
