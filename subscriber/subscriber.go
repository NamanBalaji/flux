package subscriber

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/NamanBalaji/flux/internal/subscriber/api"
	"github.com/NamanBalaji/flux/pkg/request"
)

type Subscriber struct {
	mu            sync.Mutex
	host          string
	port          int
	brokerAddress string
	topics        []string
	messageChan   chan request.PollMessage
}

func NewSubscriber(host string, port int, brokerAddr string) *Subscriber {

	sub := &Subscriber{
		host:          host,
		port:          port,
		brokerAddress: brokerAddr,
		messageChan:   make(chan request.PollMessage),
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: api.SetupRouter(sub.messageChan),
	}

	go func() {
		log.Printf("starting subscriber on port %d", port)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error occurred while trying to start the subscriber server %v \n", err)
		}
	}()

	return sub
}

func (s *Subscriber) Subscribe(topics []string, realOld bool) error {
	s.mu.Lock()
	requestBody := request.RegisterSubscriberRequest{
		Address: fmt.Sprintf("%s:%d", s.host, s.port),
		Topics:  topics,
		ReadOld: realOld,
	}
	s.mu.Unlock()

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	_, status, err := request.SendHTTPRequest(http.MethodPost, fmt.Sprintf("%s/subscrine", s.brokerAddress), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("subscriber request failed with status code %d", status)
	}

	s.mu.Lock()
	s.topics = append(s.topics, topics...)
	s.mu.Unlock()

	return nil
}

func (s *Subscriber) Unsubscribe(topics []string) error {
	s.mu.Lock()
	requestBody := request.UnsubscribeRequest{
		Address: fmt.Sprintf("%s:%d", s.host, s.port),
		Topics:  topics,
	}
	s.mu.Unlock()

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	_, status, err := request.SendHTTPRequest(http.MethodPost, fmt.Sprintf("%s/unsubscrine", s.brokerAddress), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("unsubscriber request failed with status code %d", status)
	}

	s.mu.Lock()
	for _, t := range topics {
		s.removeTopic(t)
	}
	s.mu.Unlock()

	return nil
}

func (s *Subscriber) removeTopic(topic string) {
	for i, t := range s.topics {
		if t == topic {
			s.topics = append(s.topics[:i], s.topics[i+1:]...)
		}
	}
}
