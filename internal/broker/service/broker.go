package service

import (
	"fmt"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/subscriber"
	topickPkg "github.com/NamanBalaji/flux/pkg/topic"
	"sync"
)

type Broker struct {
	mu          sync.Mutex
	Messages    message.Messages
	Topics      topickPkg.Topics
	Subscribers subscriber.SubscribersByTopic
	// other brokers
}

func NewBroker() *Broker {

	return &Broker{
		Messages:    make(message.Messages),
		Topics:      topickPkg.CreateTopics(),
		Subscribers: make(subscriber.SubscribersByTopic),
	}
}

func (b *Broker) PublishMessage(topic string, msg message.Message) bool {
	b.mu.Lock()
	// Store message
	msg.Order = b.Messages.GetLatestOrder() + 1
	b.Messages.AddMessage(msg)

	msgChan, ok := b.Topics[topic]
	if !ok {
		msgChan = make(chan message.Message, 100)
		b.Topics[topic] = msgChan
	}
	b.mu.Unlock()

	if !ok {
		go b.manageTopic(topic, msgChan)
	}

	select {
	case msgChan <- msg:
		return true
	default:
		return false
	}
}

func (b *Broker) manageTopic(topic string, msgChan chan message.Message) {
	for msg := range msgChan {
		b.deliverMessageToSubscribers(topic, msg)
	}
}

func (b *Broker) deliverMessageToSubscribers(topic string, msg message.Message) {
	b.mu.Lock()
	subscribers, ok := b.Subscribers[topic]
	if !ok {
		b.mu.Unlock()

		return
	}

	// Make a copy to safely iterate outside the lock
	subsCopy := make([]subscriber.Subscriber, len(subscribers))
	for _, sub := range subscribers {
		subsCopy = append(subsCopy, *sub)
	}
	b.mu.Unlock()

	for _, sub := range subsCopy {
		if msg.Order <= sub.StartAt {
			continue
		}
		select {
		case sub.MessageChan <- msg:
			// log that the message was enqueued successfully
		default:
			// handle error by implementing backpressure
		}
	}
}

func (b *Broker) ValidateTopics(topics []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, topic := range topics {
		_, ok := b.Topics[topic]
		if !ok {
			return fmt.Errorf("topic %s does not exist", topic)
		}
	}

	return nil
}

func (b *Broker) Subscribe(topic string, address string) {
	b.mu.Lock()
	order := b.Messages.GetLatestOrder()
	b.mu.Unlock()

	sub := subscriber.NewSubscriber(address, order)
	sub.MessageChan = make(chan message.Message, 10)

	b.mu.Lock()
	b.Subscribers.AddSubscriber(topic, sub)
	b.mu.Unlock()

	go b.processMessage(sub)

}

func (b *Broker) processMessage(subscriber *subscriber.Subscriber) {
	if subscriber == nil {
		return
	}

	for msg := range subscriber.MessageChan {
		err := subscriber.PushMessage(msg)
		if err == nil {
			subscriber.StartAt = msg.Order
		}
	}
}
