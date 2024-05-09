package message

import (
	"log"
	"sync"
	"time"

	"github.com/NamanBalaji/flux/pkg/config"
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
	log.Printf("Acked messageID %s to subscriber[address: %s] \n", m.Id, subscriberAddress)
}

func (m *Message) AddSubscriber(subscriberAddress string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.Delivered[subscriberAddress] = false
	log.Printf("messageID %s tracking ACK from subscriber[address: %s] \n", m.Id, subscriberAddress)
}

func (m *Message) RemoveSubscriber(subscriberAddress string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	delete(m.Delivered, subscriberAddress)
	log.Printf("messageID %s not tracking ACK from subscriber[address: %s] \n", m.Id, subscriberAddress)
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

	return allAcked && time.Now().Sub(m.AddedAt).Seconds() >= float64(cfg.Message.TTL)
}
