package message

import (
	"github.com/NamanBalaji/flux/pkg/config"
	"sync"
	"time"
)

type Message struct {
	Lock      sync.Mutex
	Id        string `json:"id"`
	Payload   string `json:"payload"`
	Delivered map[string]bool
	AddedAt   time.Time
}

func NewMessage(id string, payload string) *Message {
	return &Message{
		Id:        id,
		Payload:   payload,
		Delivered: make(map[string]bool),
		AddedAt:   time.Now(),
	}
}

func (m *Message) Ack(subscriberAddress string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.Delivered[subscriberAddress] = true
}

func (m *Message) AddSubscriber(subscriberAddress string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.Delivered[subscriberAddress] = false
}

func (m *Message) RemoveSubscriber(subscriberAddress string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	delete(m.Delivered, subscriberAddress)
}

func (m *Message) SafeToDelete(cfg config.Config) bool {
	allAcked := true

	m.Lock.Lock()
	defer m.Lock.Unlock()

	for _, acked := range m.Delivered {
		if !acked {
			allAcked = false

			break
		}
	}

	return allAcked && time.Now().Sub(m.AddedAt) > time.Duration(cfg.Message.TTL)*time.Second
}
