package subscriber

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/queue"
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
			err := s.pushMessage(ctx, cfg, msg, topicName)
			if err == nil {
				msg = s.MessageQueue.Dequeue()
				msg.Ack(s.Addr)
			}
		}
	}
}

func (s *Subscriber) pushMessage(ctx context.Context, cfg config.Config, msg *message.Message, topicName string) error {
	res := MessageResponse{
		Id:      msg.Id,
		Payload: msg.Payload,
		Topic:   topicName,
	}

	jsonBody, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.Subscriber.Timeout) * time.Second,
	}

	for i := 0; i < cfg.Subscriber.RetryCount; i++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context canceled: %v", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", s.Addr, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()

			return nil
		}

		if i < cfg.Subscriber.RetryCount-1 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context canceled: %v", ctx.Err())
			case <-time.After(time.Duration(cfg.Subscriber.Timeout) * time.Second):
			}
		}
	}

	s.IsActive = false
	return fmt.Errorf("failed to send message after %d retries", cfg.Subscriber.RetryCount)
}
