package subscriber

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/queue"
)

type Subscriber struct {
	Lock         sync.Mutex
	Addr         string
	MessageQueue *queue.Queue
	IsActive     bool
	CancelFunc   context.CancelFunc
	LastActive   time.Time
}

type MessageResponse struct {
	Id      string `json:"id"`
	Payload string `json:"payload"`
	Topic   string `json:"topic"`
}

func NewSubscriber(ctx context.Context, addr string) *Subscriber {
	msgQueue := queue.NewQueue()

	return &Subscriber{
		Addr:         addr,
		MessageQueue: msgQueue,
		IsActive:     true,
		LastActive:   time.Now(),
	}
}

func (s *Subscriber) AddMessage(msg *message.Message) {
	s.MessageQueue.Enqueue(msg)
	log.Printf("added messgae with id %s to subscriber[address: %s] queue \n", msg.Id, s.Addr)
}

func (s *Subscriber) HandleQueue(ctx context.Context, cfg config.Config, topicName string) {
	log.Printf("Started go routine for delivering enqueued message for subscriber[Addresss: %s] subscribed to topic %s \n", s.Addr, topicName)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Subscriber[Address: %s] subscribed to topic %s is being removed exitng out of go routine \n", s.Addr, topicName)
			return
		default:
			s.Lock.Lock()
			if !s.IsActive {
				log.Printf("Subscriber[Address: %s] subscribed to topic %s is not active returning out of go routine\n", s.Addr, topicName)
				s.Lock.Unlock()

				return
			}
			s.Lock.Unlock()

			msg := s.MessageQueue.Peek()
			err := s.pushMessage(ctx, cfg, msg, topicName)
			if err == nil {
				s.Lock.Lock()
				s.LastActive = time.Now()
				s.Lock.Unlock()

				msg = s.MessageQueue.Dequeue()
				msg.Ack(s.Addr)
			} else {
				s.Lock.Lock()
				s.IsActive = false
				s.Lock.Unlock()

				log.Printf("Subscriber[Address: %s] subscribed to topic %s encountered an error trying to push message to the subscriber: %s \n", s.Addr, topicName, err)
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

		req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/poll", s.Addr), bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		log.Printf("Sending message with id %s to subscriber[Address: %s] subscribed to topic %s \n", msg.Id, s.Addr, topicName)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()

			return nil
		} else {
			log.Printf("an error occured while trying to send request to the subscriber[Address: %s] subscribed to topic %s: err: %s\n", s.Addr, topicName, err)
		}

		if i < cfg.Subscriber.RetryCount-1 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context canceled: %v", ctx.Err())
			case <-time.After(time.Duration(cfg.Subscriber.Timeout) * time.Second):
				log.Printf("retrying sending message request to the subscriber[Address: %s] subscribed to topic %s \n", s.Addr, topicName)
			}
		}
	}

	return fmt.Errorf("failed to send message after %d retries", cfg.Subscriber.RetryCount)
}
